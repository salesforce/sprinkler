// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package table

import (
	"time"

	"gorm.io/gorm"
	"mce.salesforce.com/sprinkler/model"
)

type Workflow struct {
	gorm.Model
	Name                 string      `gorm:"type:varchar(256);not null;index:workflows_name,unique"`
	Artifact             string      `gorm:"type:varchar(2048);not null"`
	Command              string      `gorm:"type:text;not null"`
	Every                model.Every `gorm:"type:varchar(64);not null"`
	NextRuntime          time.Time   `gorm:"not null"`
	Backfill             bool        `gorm:"not null"`
	Owner                *string     `gorm:"type:varchar(2048)"`
	IsActive             bool        `gorm:"not null"`
	ScheduleDelayMinutes uint        `gorm:"default:0"`

	ScheduledWorkflows []ScheduledWorkflow
}

type ScheduledWorkflow struct {
	gorm.Model
	WorkflowID         uint
	OrchardID          string    `gorm:"type:varchar(64);not null"`
	StartTime          time.Time `gorm:"not null"`
	ScheduledStartTime time.Time `gorm:"not null"`
	Status             string    `gorm:"type:varchar(64);not null"`
}

type WorkflowSchedulerLock struct {
	WorkflowID uint      `gorm:"primaryKey"`
	Token      string    `gorm:"type:varchar(64);not null"`
	LockTime   time.Time `gorm:"not null"`
}

type WorkflowActivatorLock struct {
	ScheduledID uint      `gorm:"primaryKey"`
	Token       string    `gorm:"type:varchar(64);not null"`
	LockTime    time.Time `gorm:"not null"`
}
