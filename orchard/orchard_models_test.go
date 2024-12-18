// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package orchard

import (
	"fmt"
	"testing"
)

func TestDetails(t *testing.T) {
	wfId := "wf-e7901f9f-8e88-4173-9d6e-d63f8b1e647f"
	wfName := "test.workflow"
	wfStats := "pending"
	wfCreatedAt := "2024-12-12T02:12:19.260463"
	wfStr := fmt.Sprintf(`{"id":"%s","name":"%s","status":"%s","createdAt":"%s"}`, wfId, wfName, wfStats, wfCreatedAt)
	resp := []byte(wfStr)
	details, err := ParseDetails(resp)
	if err != nil {
		t.Fatalf("unable to parse response: %v", err)
	}
	if details.Id != wfId {
		t.Fatalf("parsed %s doesn't match %s", details.Id, wfId)
	}
	if details.Name != wfName {
		t.Fatalf("parsed %s doesn't match %s", details.Name, wfName)
	}
	if details.Status != wfStats {
		t.Fatalf("parsed %s doesn't match %s", details.Status, wfStats)
	}
	if details.CreatedAt != wfCreatedAt {
		t.Fatalf("parsed %s doesn't match %s", details.CreatedAt, wfCreatedAt)
	}
}
