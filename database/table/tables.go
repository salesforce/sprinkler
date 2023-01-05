package table

import (
	"time"

	"gorm.io/gorm"
)

type Workflow struct {
	gorm.Model
	Name        string    `gorm:"type:varchar(256);not null"`
	Artifact    string    `gorm:"type:varchar(2048);not null"`
	Command     string    `gorm:"type:text;not null"`
	Every       string    `gorm:"type:varchar(64);not null"`
	NextRuntime time.Time `gorm:"not null"`
	Backfill    bool      `gorm:"not null"`
	Owner       *string   `gorm:"type:varchar(2048)"`
	IsActive    bool      `gorm:"not null"`

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
