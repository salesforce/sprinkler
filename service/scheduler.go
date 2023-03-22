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
	"mce.salesforce.com/sprinkler/model"
	"mce.salesforce.com/sprinkler/orchard"
)

type Scheduler struct {
	Interval    time.Duration
	MaxSize     uint
	OrchardHost string
}

func (s *Scheduler) Start() {
	fmt.Println("Scheduler Started")
	tick := time.Tick(s.Interval)
	for range tick {
		fmt.Println("tick")
		s.scheduleWorkflows(database.GetInstance())
	}
}

func (s *Scheduler) scheduleWorkflows(db *gorm.DB) {
	var workflows []table.Workflow

	db.Model(&table.Workflow{}).
		Joins("left join workflow_scheduler_locks l on workflows.id = l.workflow_id").
		Where("next_runtime <= ? and l.token is null", time.Now()).
		Find(&workflows)

	// db.Where("next_runtime <= ?", time.Now()).Find(&workflows)
	for _, wf := range workflows {
		go s.lockAndRun(db, wf)
	}
}

func (s *Scheduler) lockAndRun(db *gorm.DB, wf table.Workflow) {
	token := uuid.New().String()

	lock := table.WorkflowSchedulerLock{
		wf.ID,
		token,
		time.Now(),
	}
	result := db.Create(&lock)
	if result.Error != nil {
		fmt.Printf("something else is running this workflow %v! skip...\n", wf.ID)
		return
	}

	existingLock := table.WorkflowSchedulerLock{}
	db.First(&existingLock, wf.ID)

	if existingLock.Token != token {
		fmt.Printf("something else is running this workflow %v! skip...\n", wf.ID)
		return
	}

	fmt.Println("running workflow", wf.Name, token)
	client := &orchard.OrchardRestClient{
		Host:   s.OrchardHost,
		Runner: orchard.OrchardJavaRunner{},
	}
	orchardID, err := client.Create(wf)
	if err != nil {
		// TODO should not panic here
		fmt.Println(err)
		panic("Don't panic, do something1")
	}
	err = client.Activate(orchardID)
	scheduleStatus := "activated"
	if err != nil {
		fmt.Println(err)
		scheduleStatus = "error"
	}

	// add to scheduled and update the next run time
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&table.ScheduledWorkflow{
			WorkflowID:         wf.ID,
			OrchardID:          orchardID,
			StartTime:          time.Now(),
			ScheduledStartTime: wf.NextRuntime,
			Status:             scheduleStatus,
		}).Error; err != nil {
			return err
		}

		fmt.Println(wf.Every)

		if err := tx.Model(&wf).Update(
			"next_runtime", nextRuntime(wf.NextRuntime, wf.Every, wf.Backfill),
		).Error; err != nil {
			return err
		}
		return nil
	})

	// release the lock
	db.Where("workflow_id = ? and token = ?", wf.ID, token).
		Delete(&table.WorkflowSchedulerLock{})
}

// ignore addInterval parsing error here, since it shouldn't fail
func nextRuntime(start time.Time, every model.Every, backfill bool) time.Time {
	next := addInterval(start, every)
	if backfill || next.After(time.Now()) {
		return next
	} else {
		return nextRuntime(next, every, backfill)
	}
}

func addInterval(someTime time.Time, every model.Every) time.Time {
	switch every.Unit {
	case model.EveryMinute:
		return someTime.Add(time.Duration(every.Quantity) * time.Minute)
	case model.EveryDay:
		return someTime.AddDate(0, 0, int(every.Quantity))
	case model.EveryWeek:
		return someTime.AddDate(0, 0, int(every.Quantity*7))
	case model.EveryMonth:
		return someTime.AddDate(0, int(every.Quantity), 0)
	case model.EveryYear:
		return someTime.AddDate(int(every.Quantity), 0, 0)
	}
	panic("wrong")
}
