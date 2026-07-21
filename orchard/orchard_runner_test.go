// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package orchard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestProcessCmdConcurrentWorkingDir guards against a process-global working
// directory race: concurrent callers each expect their command to run in their
// own directory. A per-directory marker file is read back via `cat`; if one
// goroutine's directory change leaked into another, the wrong file (or no file)
// would be read and the assertion would fail.
func TestProcessCmdConcurrentWorkingDir(t *testing.T) {
	const goroutines = 25

	dirs := make([]string, goroutines)
	for i := range dirs {
		dir := t.TempDir()
		// unique payload per dir, emitted as a JSON array so processCmd parses it
		payload := fmt.Sprintf("dir-%d", i)
		content, _ := json.Marshal([]string{payload})
		if err := os.WriteFile(filepath.Join(dir, "marker.json"), content, 0o600); err != nil {
			t.Fatalf("writing marker: %v", err)
		}
		dirs[i] = dir
	}

	var wg sync.WaitGroup
	errs := make([]error, goroutines)
	got := make([][]string, goroutines)
	for i := range dirs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			out, err := processCmd(`["cat", "marker.json"]`, dirs[i])
			got[i], errs[i] = out, err
		}(i)
	}
	wg.Wait()

	for i := range dirs {
		if errs[i] != nil {
			t.Errorf("goroutine %d: unexpected error: %v", i, errs[i])
			continue
		}
		want := fmt.Sprintf("dir-%d", i)
		if len(got[i]) != 1 || got[i][0] != want {
			t.Errorf("goroutine %d: read %v from wrong working dir, want [%q]", i, got[i], want)
		}
	}
}
