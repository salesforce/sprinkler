// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package model

import (
	"testing"
)

func TestEvery(t *testing.T) {
	str := "1.minute"

	every, err := ParseEvery(str)

	if err != nil {
		t.Fatalf("got error: %v with input %q", err, str)
	}

	if every.Quantity != 1 {
		t.Fatalf("quantity is not 1 with input: %q", str)
	}

	str = "1.second"
	every, err = ParseEvery(str)

	if err == nil {
		t.Fatalf("should not succeed with input: %q", str)
	}

	str = "somerandomstuff1213"
	every, err = ParseEvery(str)
	if err == nil {
		t.Fatalf("should not succeed with input: %q", str)
	}

}
