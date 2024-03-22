package service

import (
	"fmt"
	"gorm.io/gorm"
	"mce.salesforce.com/sprinkler/database"
	"mce.salesforce.com/sprinkler/database/table"
	"time"
)

type Cleanup struct {
	ScheduledWorkflowTimeout      time.Duration
	WorkflowActivationLockTimeout time.Duration
	WorkflowSchedulerLockTimeout  time.Duration
}

func (s *Cleanup) Run() {
	s.deleteExpiredActivatorLocks(database.GetInstance())
	fmt.Println("Deleting expired scheduler locks...")
	s.deleteExpiredSchedulerLocks(database.GetInstance())
	fmt.Println("Deleting expired scheduled workflows...")
	s.deleteExpiredScheduledWorkflows(database.GetInstance())
	fmt.Println("Cleanup complete")
}

func (s *Cleanup) deleteExpiredActivatorLocks(db *gorm.DB) {
	expiryTime := time.Now().Add(-s.WorkflowActivationLockTimeout)
	fmt.Printf("Deleting activation locks older than $s ...\n", expiryTime)

	db.Model(&table.WorkflowActivatorLock{}).
		Where("lock_time < ?", expiryTime).
		Delete(&table.WorkflowActivatorLock{})
}

func (s *Cleanup) deleteExpiredSchedulerLocks(db *gorm.DB) {
	expiryTime := time.Now().Add(-s.WorkflowSchedulerLockTimeout)
	fmt.Printf("Deleting scheduler locks older than $s ...\n", expiryTime)

	db.Model(&table.WorkflowSchedulerLock{}).
		Where("lock_time < ?", expiryTime).
		Delete(&table.WorkflowSchedulerLock{})
}

func (s *Cleanup) deleteExpiredScheduledWorkflows(db *gorm.DB) {
	expiryTime := time.Now().Add(-s.ScheduledWorkflowTimeout)
	fmt.Printf("Deleting scheduled workflows older than $s ...\n", expiryTime)

	db.Unscoped().Model(&table.ScheduledWorkflow{}).
		Where("updated_at < ?", expiryTime).
		Delete(&table.ScheduledWorkflow{})
}
