// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package orchard

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"mce.salesforce.com/sprinkler/common"
)

type OrchardJavaRunner struct{}

func (r OrchardJavaRunner) Generate(artifact string, command string) (string, error) {
	javaParts, err := r.Parse(artifact, command)
	if err != nil {
		return "", err
	}
	s3c, err := common.DefaultS3Client()
	if err != nil {
		return "", err
	}
	bb := common.BucketBasics{S3Client: s3c}
	err = bb.DownloadFile(javaParts.Bucket, javaParts.Path, javaParts.Jar)
	if err != nil {
		return "", err
	}
	cmds := []string{"java", "-cp", javaParts.Jar, javaParts.MainClass}
	if javaParts.Args != "" {
		for _, arg := range strings.Split(javaParts.Args, " ") {
			arg = strings.ReplaceAll(arg, " ", "")
			if arg != "" {
				cmds = append(cmds, arg)
			}
		}
	}
	log.Println("executing reconstructed command:", cmds)
	out, err := exec.Command(cmds[0], cmds[1:]...).Output()

	// clean up downloaded jar file
	defer func() {
		_, err2 := exec.Command("rm", javaParts.Jar).Output()
		if err2 != nil {
			log.Println("local jar cleanup error:", err2)
		} else {
			log.Println("removed downloaded local jar ", javaParts.Jar)
		}
	}()

	if err != nil {
		return "", err
	}
	return string(out), nil
}

type JavaParts struct {
	Jar       string
	MainClass string
	Args      string
	Bucket    string
	Path      string
}

// parse artifact expecting a s3 URI to return the bucket and path
// expects command `java -cp localJarFileName.jar path.to.mainClass optionalArgs` to generate Orchard create payload
func (r OrchardJavaRunner) Parse(artifact string, command string) (JavaParts, error) {
	aPatt, err := regexp.Compile(`^s3://([^ /]+)/([^ ]+)$`)
	if err != nil {
		return JavaParts{}, err
	}
	aRes := aPatt.FindStringSubmatch(artifact)
	if len(aRes) != 3 {
		return JavaParts{}, fmt.Errorf("artifact parsing failed")
	}

	jPatt, err := regexp.Compile(`^java +\-cp +([A-Za-z0-9-_./]+\.jar) +([A-Za-z0-9-_$.]+) *([A-Za-z0-9 -]*)$`)
	if err != nil {
		return JavaParts{}, err
	}
	jRes := jPatt.FindStringSubmatch(command)
	if len(jRes) != 4 {
		return JavaParts{}, fmt.Errorf("command parsing failed")
	}

	return JavaParts{jRes[1], jRes[2], jRes[3], aRes[1], aRes[2]}, nil
}
