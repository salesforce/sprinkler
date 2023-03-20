// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package orchard

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type OrchardRunner interface {
	Generate(artifact string, command string) string
}

type OrchardStdoutRunner struct{}

func (self OrchardStdoutRunner) Generate(artifact string, command string) (string, error) {
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
