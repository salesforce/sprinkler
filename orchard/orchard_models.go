// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package orchard

import (
	"encoding/json"
	"fmt"
)

type Details struct {
	Id        string `json:"id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Status    string `json:"status" binding:"required"`
	CreatedAt string `json:"createdAt" binding:"required"`
}

func ParseDetails(resp []byte) (*Details, error) {
	var data Details
	err := json.Unmarshal(resp, &data)
	if err != nil {
		return nil, fmt.Errorf("unable to decode the response: %v", err)
	}
	return &data, nil
}
