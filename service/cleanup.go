package service

import (
	"fmt"
	"gorm.io/gorm"
	"mce.salesforce.com/sprinkler/database"
	"mce.salesforce.com/sprinkler/database/table"
	"time"
)

type Cleanup struct {
	ScheduledWorkflowTimeout time.Duration
}

func (s *Cleanup) Run() {
	fmt.Println("Cleanup started")
	s.deleteExpiredScheduledWorkflows(database.GetInstance())
	fmt.Println("Cleanup complete")
}

func (s *Cleanup) deleteExpiredScheduledWorkflows(db *gorm.DB) {
	expiryTime := time.Now().Add(-s.ScheduledWorkflowTimeout)

	db.Unscoped().
		Model(&table.ScheduledWorkflow{}).
		Where("updated_at < ?", expiryTime).
		Delete(&table.ScheduledWorkflow{})
}
