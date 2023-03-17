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

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Wipe clean all the tables",
	Long:  `Drop all the sprinkler tables from the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		destroyDatabase()
	},
}

func destroyDatabase() {
	runDatabaseEffects(func(db *gorm.DB) {
		size := len(database.Tables)
		reverseTables := make([]interface{}, size)
		for i, t := range database.Tables {
			reverseTables[size-1-i] = t
		}
		db.Migrator().DropTable(reverseTables...)
	})
}

func init() {
	databaseCmd.AddCommand(destroyCmd)
}
