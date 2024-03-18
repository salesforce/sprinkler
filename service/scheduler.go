// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mce.salesforce.com/sprinkler/common"
	"mce.salesforce.com/sprinkler/database"
	"mce.salesforce.com/sprinkler/database/table"
	"mce.salesforce.com/sprinkler/model"
	"mce.salesforce.com/sprinkler/orchard"
)

// https://docs.aws.amazon.com/sns/latest/api/API_Publish.html
const SNSMessageMaxBytes int = 262144

type ScheduleStatus int

const (
	Canceled ScheduleStatus = iota
	CancelFailed
	Deleted
	DeleteFailed
	Activated
	Created
)

func (s ScheduleStatus) ToString() string {
	switch s {
	case Canceled:
		return "canceled"
	case CancelFailed:
		return "cancel_failed"
	case Deleted:
		return "deleted"
	case DeleteFailed:
		return "delete_failed"
	case Activated:
		return "activated"
	case Created:
		return "created"
	}
	panic("unknown ScheduleStatus")
}

type Scheduler struct {
	Interval                 time.Duration
	LockTimeout              time.Duration
	MaxSize                  uint
	OrchardHost              string
	OrchardAPIKeyName        string
	OrchardAPIKey            string
	ScheduledWorkflowTimeout time.Duration
}

func (s *Scheduler) Start() {
	fmt.Println("Scheduler Started")
	tick := time.Tick(s.Interval)
	for range tick {
		fmt.Println("tick")
		s.deleteExpiredLocks(database.GetInstance())
		s.scheduleWorkflows(database.GetInstance())
		s.activateWorkflows(database.GetInstance())
	}
}

func (s *Scheduler) deleteExpiredLocks(db *gorm.DB) {
	expiryTime := time.Now().Add(-s.LockTimeout)

	db.Model(&table.WorkflowActivatorLock{}).
		Where("lock_time < ?", expiryTime).
		Delete(&table.WorkflowActivatorLock{})

	db.Model(&table.WorkflowSchedulerLock{}).
		Where("lock_time < ?", expiryTime).
		Delete(&table.WorkflowSchedulerLock{})
}

func (s *Scheduler) deleteExpiredScheduledWorkflows(db *gorm.DB) {
	expiryTime := time.Now().Add(-s.ScheduledWorkflowTimeout)

	db.Model(&table.ScheduledWorkflow{}).
		Where("updated_at < ?", expiryTime).
		Delete(&table.ScheduledWorkflow{})
}

func (s *Scheduler) scheduleWorkflows(db *gorm.DB) {
	var workflows []table.Workflow

	db.Model(&table.Workflow{}).
		Joins("left join workflow_scheduler_locks l on workflows.id = l.workflow_id").
		Where("next_runtime <= ? and is_active = 't' and l.token is null", time.Now()).
		Find(&workflows)

	for _, wf := range workflows {
		go s.lockAndCreate(db, wf)
	}
}

func (s *Scheduler) activateWorkflows(db *gorm.DB) {
	var scheduledWorkflows []table.ScheduledWorkflow

	db.Model(&table.ScheduledWorkflow{}).
		Joins("left join workflow_activator_locks l on scheduled_workflows.id = l.scheduled_id").
		Where("start_time <= ? and status = 'created' and l.token is null", time.Now()).
		Order("start_time").
		Find(&scheduledWorkflows)

	for _, swf := range scheduledWorkflows {
		go s.lockAndActivate(db, swf)
	}
}

func (s *Scheduler) deleteWorkflows(
	client *orchard.OrchardRestClient,
	workflowIDs []string,
	statuses map[string]string,
) map[string]string {
	updatedStatuses := statuses
	for _, orchardID := range workflowIDs {
		// delete created workflows
		if err := client.Delete(orchardID); err != nil {
			fmt.Printf("[error] error deleting workflow: %s\n", err)
			updatedStatuses[orchardID] = DeleteFailed.ToString()
		} else {
			updatedStatuses[orchardID] = Deleted.ToString()
		}
	}
	return updatedStatuses
}

func (s *Scheduler) cancelWorkflows(
	client *orchard.OrchardRestClient,
	statuses map[string]string,
) map[string]string {
	updatedStatuses := statuses
	for orchardID, status := range statuses {
		if status == Activated.ToString() {
			// cancel activated workflows
			if err := client.Cancel(orchardID); err != nil {
				fmt.Printf("[error] error canceling workflow: %s\n", err)
				updatedStatuses[orchardID] = CancelFailed.ToString()
			} else {
				updatedStatuses[orchardID] = Canceled.ToString()
			}
		}
	}
	return updatedStatuses
}

func (s *Scheduler) createWorkflow(
	client *orchard.OrchardRestClient,
	wf table.Workflow,
) map[string]string {
	statuses := make(map[string]string)
	createdIDs, err := client.Create(wf)
	if err != nil {
		fmt.Printf("[error] error creating workflow: %s\n", err)
		notifyOwner(wf, err)
		return s.deleteWorkflows(client, createdIDs, statuses)
	}
	for _, createdID := range createdIDs {
		statuses[createdID] = Created.ToString()
	}
	return statuses
}

func (s *Scheduler) activateWorkflow(
	client *orchard.OrchardRestClient,
	swf table.ScheduledWorkflow,
	wf table.Workflow,
) string {
	if err := client.Activate(swf.OrchardID); err != nil {
		fmt.Printf("[error] error activating workflow: %s\n", err)
		notifyOwner(wf, err)
		return swf.Status
	}
	return Activated.ToString()
}

func (s *Scheduler) lockAndCreate(db *gorm.DB, wf table.Workflow) {
	token := uuid.New().String()

	lock := table.WorkflowSchedulerLock{
		wf.ID,
		token,
		time.Now(),
	}
	result := db.Create(&lock)
	if result.Error != nil {
		fmt.Printf("something else is creating this workflow %v! skip...\n", wf.ID)
		return
	}

	existingLock := table.WorkflowSchedulerLock{}
	db.First(&existingLock, wf.ID)

	if existingLock.Token != token {
		fmt.Printf("something else is creating this workflow %v! skip...\n", wf.ID)
		return
	}

	// release the lock
	defer db.Where("workflow_id = ? and token = ?", wf.ID, token).
		Delete(&table.WorkflowSchedulerLock{})

	fmt.Println("creating workflow", wf.Name, token)
	client := &orchard.OrchardRestClient{
		Host:       s.OrchardHost,
		APIKeyName: s.OrchardAPIKeyName,
		APIKey:     s.OrchardAPIKey,
	}

	scheduleStatus := s.createWorkflow(client, wf)

	// add to scheduled and update the next run time
	db.Transaction(func(tx *gorm.DB) error {
		startTime := time.Now()
		for orchardID, status := range scheduleStatus {
			if err := tx.Create(&table.ScheduledWorkflow{
				WorkflowID:         wf.ID,
				OrchardID:          orchardID,
				StartTime:          startTime,
				ScheduledStartTime: wf.NextRuntime,
				Status:             status,
			}).Error; err != nil {
				return err
			}
			startTime = startTime.Add(time.Duration(wf.ScheduleDelayMinutes) * time.Minute)
		}

		fmt.Println(wf.Every)

		if err := tx.Model(&wf).Update(
			"next_runtime", nextRuntime(wf.NextRuntime, wf.Every, wf.Backfill),
		).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *Scheduler) lockAndActivate(db *gorm.DB, swf table.ScheduledWorkflow) {
	token := uuid.New().String()

	lock := table.WorkflowActivatorLock{
		swf.ID,
		token,
		time.Now(),
	}
	result := db.Create(&lock)
	if result.Error != nil {
		fmt.Printf("something else is activating this scheduled workflow %v! skip...\n", swf.ID)
		return
	}

	existingLock := table.WorkflowActivatorLock{}
	db.First(&existingLock, swf.ID)

	if existingLock.Token != token {
		fmt.Printf("something else is activating this scheduled workflow %v! skip...\n", swf.ID)
		return
	}

	// release the lock
	defer db.Where("scheduled_id = ? and token = ?", swf.ID, token).
		Delete(&table.WorkflowActivatorLock{})

	fmt.Println("activating workflow", swf.OrchardID, token)
	client := &orchard.OrchardRestClient{
		Host:       s.OrchardHost,
		APIKeyName: s.OrchardAPIKeyName,
		APIKey:     s.OrchardAPIKey,
	}

	wf := table.Workflow{}
	db.First(&wf, swf.WorkflowID)

	status := s.activateWorkflow(client, swf, wf)

	// update status in scheduled_workflows table
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&swf).Update("status", status).Error; err != nil {
			return err
		}
		return nil
	})
}

func notifyOwner(wf table.Workflow, orchardErr error) {
	errMsg := fmt.Sprintf(
		"[error] Failed to schedule workflow %v with error %q\n",
		wf,
		orchardErr,
	)
	log.Println(errMsg)
	if wf.Owner == nil || *wf.Owner == "" {
		return
	}

	cred := common.WithAwsCredentials()
	client, err := cred.SNSClient()
	if err != nil {
		log.Println("[error] error initiating SNS client")
	}

	croppedErrMsg := truncate(errMsg, SNSMessageMaxBytes)
	input := &sns.PublishInput{
		Message:  &croppedErrMsg,
		TopicArn: wf.Owner,
	}

	result, err := client.Publish(
		context.TODO(),
		input,
	)

	if err != nil {
		log.Println("[error] error publishing SNS message")
		log.Println(err)
		return
	}

	log.Printf(
		"Successful notify owner %q, with message ID: %q\n",
		*wf.Owner,
		*result.MessageId,
	)
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
	case model.EveryHour:
		return someTime.Add(time.Duration(every.Quantity) * time.Hour)
	case model.EveryDay:
		return someTime.AddDate(0, 0, int(every.Quantity))
	case model.EveryWeek:
		return someTime.AddDate(0, 0, int(every.Quantity*7))
	case model.EveryMonth:
		return someTime.AddDate(0, int(every.Quantity), 0)
	case model.EveryYear:
		return someTime.AddDate(int(every.Quantity), 0, 0)
	}
	panic(fmt.Sprintf("Every unit '%s' not recognized", every.Unit))
}

func truncate(str string, bytes int) string {
	slice := []byte(str)
	if len(slice) <= bytes {
		return str
	}
	return string(slice[:bytes])
}
