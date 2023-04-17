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

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the database tables",
	Long:  `Migrate all tables for sprinkler.`,
	Run: func(cmd *cobra.Command, args []string) {
		migrateDatabase()
	},
}

func migrateDatabase() {
	runDatabaseEffects(func(db *gorm.DB) {
		if db.AutoMigrate(database.Tables...) != nil {
			panic("Failed creating tables")
		}
	})
}

func init() {
	databaseCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().BoolVar(
		&withSample, "with-sample", false, "With Sample Data")
}
