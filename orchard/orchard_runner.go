// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package orchard

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"mce.salesforce.com/sprinkler/common"
)

const baseDir string = "/sprinkler"

type OrchardRunner interface {
	Generate(artifact string, command string) (string, error)
}

type OrchardStdoutRunner struct{}

func (r OrchardStdoutRunner) Generate(artifact string, command string) (string, error) {
	if artifact == "" {
		return processCmd(command, baseDir)
	}

	if !strings.HasPrefix(artifact, "s3://") {
		return "", fmt.Errorf("artifact %v is not supported\n", artifact)
	}

	return s3ArtifactGenerate(artifact, command)
}

func s3ArtifactGenerate(artifact string, command string) (string, error) {

	// tmp directory to avoid threads race on downloaded artifact
	tmpDir, err := os.MkdirTemp("", "sprinkler-")
	if err != nil {
		return "", fmt.Errorf("problem creating a tmp directory: %w", err)
	}

	// download s3 artifact as local file in tmp directory
	s3c, err := common.DefaultS3Client()
	if err != nil {
		return "", err
	}
	s3 := common.S3Basics{S3Client: s3c}
	s3bucketPath, err := s3.GetBucketPath(artifact)
	if err != nil {
		return "", err
	}
	localFile, err := s3.GetLastSegment(s3bucketPath.Path)
	if err != nil {
		return "", fmt.Errorf("problem extracting filename from artifact path: %w", err)
	}
	localFile = tmpDir + "/" + localFile
	err = s3.DownloadFile(s3bucketPath.Bucket, s3bucketPath.Path, localFile)
	if err != nil {
		return "", fmt.Errorf("problem downloading s3 artifact %v to %v: %w", s3bucketPath.Path, localFile, err)
	}
	log.Printf("downloaded %v\n", localFile)

	// clean up downloaded jar file
	defer func(tmpDir string) {
		if err2 := os.RemoveAll(tmpDir); err2 != nil {
			log.Printf("local downloaded artifact cleanup error:%v\n", err2)
		} else {
			log.Printf("finished cleanup downloaded file %v\n", localFile)
		}
	}(tmpDir)

	return processCmd(command, tmpDir)
}

func processCmd(command string, pwd string) (string, error) {
	if err := os.Chdir(pwd); err != nil {
		return "", fmt.Errorf("cd %v has error:%w", pwd, err)
	}
	cmds, err := parseCommandLine(command)
	if err != nil {
		return "", fmt.Errorf("parse command line error%w", err)
	}
	if len(cmds) < 1 {
		return "", fmt.Errorf("Invalid command line %s", command)
	}
	out, err := exec.Command(cmds[0], cmds[1:]...).Output()
	if err != nil {
		return "", fmt.Errorf("exec command %v has error:%w", command, err)
	}
	return string(out), nil
}

func parseCommandLine(command string) ([]string, error) {
	var output []string
	err := json.Unmarshal([]byte(command), &output)
	return output, err
}
