// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"mce.salesforce.com/sprinkler/database"
	"mce.salesforce.com/sprinkler/database/table"
)

var withSample = false

// initializeCmd represents the initialize command
var initializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Initialize the database tables",
	Long:  `Create and setup all tables for sprinkler.`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeDatabase()
	},
}

func initializeDatabase() {
	runDatabaseEffects(func(db *gorm.DB) {
		if db.AutoMigrate(&table.Workflow{}, &table.ScheduledWorkflow{}, &table.WorkflowSchedulerLock{}) != nil {
			panic("Failed creating tables")
		}

		if db.AutoMigrate(&table.SchemaEvolution{}) != nil {
			panic("Failed creating SchemaEvolution")
		}
		v1 := db.Where("schema_version = ?", "1").First(&table.SchemaEvolution{})
		if v1.Error != nil && !errors.Is(v1.Error, gorm.ErrRecordNotFound) {
			panic("Failed checking SchemaVersion 1")
		}
		if v1.RowsAffected != 1 {
			// record created table.SchemaEvolution
			db.Create(&table.SchemaEvolution{SchemaVersion: "1", ExecutedAt: time.Now()})
		}

		v2 := db.Where("schema_version = ?", "2").First(&table.SchemaEvolution{})
		if v2.Error != nil && !errors.Is(v2.Error, gorm.ErrRecordNotFound) {
			panic("Failed checking SchemaVersion 2")
		}
		if v2.RowsAffected != 1 {
			db.Migrator().CreateIndex(&table.Workflow{}, "Name")
			db.Migrator().CreateIndex(&table.Workflow{}, "Artifact")
			db.Migrator().CreateIndex(&table.Workflow{}, "workflows_name_artifact")
			// record created workflows_name_artifact unique index
			db.Create(&table.SchemaEvolution{SchemaVersion: "2", ExecutedAt: time.Now()})
		}

		if withSample {
			db.Create(&database.SampleWorkflows)
		}
	})
}

func init() {
	databaseCmd.AddCommand(initializeCmd)

	initializeCmd.Flags().BoolVar(
		&withSample, "with-sample", false, "With Sample Data")
}
