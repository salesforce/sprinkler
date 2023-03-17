// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"mce.salesforce.com/sprinkler/database"
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
		if db.Migrator().CreateTable(database.Tables...) != nil {
			panic("Failed creating tables")
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
