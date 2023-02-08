// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"mce.salesforce.com/sprinkler/database"
	"mce.salesforce.com/sprinkler/database/table"
	"mce.salesforce.com/sprinkler/orchard"
)

type Scheduler struct {
	Interval time.Duration
	MaxSize  uint
}

func (s *Scheduler) Start() {
	fmt.Println("Scheduler Started")
	tick := time.Tick(s.Interval)
	for _ = range tick {
		fmt.Println("tick")
		scheduleWorkflows(database.GetInstance())
	}
}

func scheduleWorkflows(db *gorm.DB) {
	var workflows []table.Workflow

	db.Model(&table.Workflow{}).
		Joins("left join workflow_scheduler_locks l on workflows.id = l.workflow_id").
		Where("next_runtime <= ? and l.token is null", time.Now()).
		Find(&workflows)

	// db.Where("next_runtime <= ?", time.Now()).Find(&workflows)
	for _, wf := range workflows {
		go lockAndRun(db, wf)
	}
}

func lockAndRun(db *gorm.DB, wf table.Workflow) {
	token := uuid.New().String()

	lock := table.WorkflowSchedulerLock{
		wf.ID,
		token,
		time.Now(),
	}
	result := db.Create(&lock)
	if result.Error != nil {
		fmt.Printf("something else is running this workflow %s! skip...\n", wf.ID)
		return
	}

	existingLock := table.WorkflowSchedulerLock{}
	db.First(&existingLock, wf.ID)

	if existingLock.Token != token {
		fmt.Println("something else is running this workflow %s! skip...", wf.ID)
		return
	}

	fmt.Println("running workflow", wf.Name, token)
	client := &orchard.OrchardRestClient{
		Host: "http://ws:8080",
	}
	orchardID, err := client.Create(wf)
	if err != nil {
		// TODO should not panic here
		fmt.Println(err)
		panic("Don't panic, do something1")
	}
	err = client.Activate(orchardID)
	if err != nil {
		// TODO should not panic here
		fmt.Println(err)
		panic("Dont' panic, do something2")
	}

	// release the lock
	db.Where("workflow_id = ? and token = ?", wf.ID, token).
		Delete(&table.WorkflowSchedulerLock{})
}
