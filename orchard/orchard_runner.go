// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package orchard

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"mce.salesforce.com/sprinkler/common"
)

type OrchardRunner interface {
	Generate(artifact string, command string) (string, error)
}

type OrchardStdoutRunner struct{}

func (r OrchardStdoutRunner) Generate(artifact string, command string) (string, error) {
	if artifact == "" {
		return processCmd(command)
	}

	if !strings.HasPrefix(artifact, "s3://") {
		return "", fmt.Errorf("artifact %v is not supported\n", artifact)
	}

	// download s3 artifact as local file
	s3c, err := common.DefaultS3Client()
	bb := common.S3Basics{S3Client: s3c}
	s3bucketPath, err := bb.GetBucketPath(artifact)
	localFile, err := bb.GetLastSegment(s3bucketPath.Path)
	err = bb.DownloadFile(s3bucketPath.Bucket, s3bucketPath.Path, localFile)
	if err != nil {
		return "", nil
	}

	// clean up downloaded jar file
	defer func() {
		_, err2 := exec.Command("rm", localFile).Output()
		if err2 != nil {
			log.Println("local jar cleanup error:", err2)
		} else {
			log.Println("removed downloaded local file:", localFile)
		}
	}()

	return processCmd(command)
}

func processCmd(command string) (string, error) {
	cmds, err := parseCommandLine(command)
	if err != nil {
		return "", err
	}
	if len(cmds) < 1 {
		return "", fmt.Errorf("Invalid command line %s", command)
	}
	out, err := exec.Command(cmds[0], cmds[1:]...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func parseCommandLine(command string) ([]string, error) {
	var output []string
	err := json.Unmarshal([]byte(command), &output)
	return output, err
}
