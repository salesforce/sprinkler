// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package database

import (
	"os"
	"time"

	"mce.salesforce.com/sprinkler/database/table"
	"mce.salesforce.com/sprinkler/model"
)

var Tables = []interface{}{
	&table.Workflow{},
	&table.ScheduledWorkflow{},
	&table.WorkflowSchedulerLock{},
	&table.WorkflowActivatorLock{},
}

var owner = os.Getenv("OWNER_SNS")

var SampleWorkflows = []table.Workflow{
	table.Workflow{
		Name:        "1",
		Artifact:    "", // empty string means local available
		Command:     `["echo", "{\"name\": \"command1\"}"]`,
		Every:       model.Every{1, model.EveryDay},
		NextRuntime: time.Now().Add(-1 * time.Hour),
		Backfill:    false,
		Owner:       &owner,
		IsActive:    true,
	},
	table.Workflow{
		Name:        "2",
		Artifact:    "",
		Command:     `["echo", "{\"name\": \"command2\"}"]`,
		Every:       model.Every{1, model.EveryDay},
		NextRuntime: time.Now().Add(24 * time.Hour),
		Backfill:    false,
		Owner:       &owner,
		IsActive:    true,
	},
	table.Workflow{
		Name:        "3",
		Artifact:    "",
		Command:     `["echo", "{\"name\": \"command3\"}"]`,
		Every:       model.Every{1, model.EveryDay},
		NextRuntime: time.Now().Add(-24 * time.Hour),
		Backfill:    true,
		Owner:       &owner,
		IsActive:    true,
	},
}
