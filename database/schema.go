// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package database

import (
	"time"

	"mce.salesforce.com/sprinkler/database/table"
)

var Tables = []interface{}{
	&table.Product{},
	&table.Workflow{},
	&table.ScheduledWorkflow{},
	&table.WorkflowSchedulerLock{},
}

var owner = "example@example.com"

var SampleWorkflows = []table.Workflow{
	table.Workflow{
		Name:        "1",
		Artifact:    "", // empty string means local available
		Command:     `["echo", "{\"name\": \"command1\"}"]`,
		Every:       "1.day",
		NextRuntime: time.Now(),
		Backfill:    false,
		Owner:       &owner,
		IsActive:    true,
	},
	table.Workflow{
		Name:        "2",
		Artifact:    "",
		Command:     `["echo", "{\"name\": \"command2\"}"]`,
		Every:       "1.day",
		NextRuntime: time.Now().Add(24 * time.Hour),
		Backfill:    false,
		Owner:       &owner,
		IsActive:    true,
	},
	table.Workflow{
		Name:        "3",
		Artifact:    "",
		Command:     `["echo", "{\"name\": \"command3\"}"]`,
		Every:       "1.day",
		NextRuntime: time.Now().Add(-24 * time.Hour),
		Backfill:    false,
		Owner:       &owner,
		IsActive:    true,
	},
}
