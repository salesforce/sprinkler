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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
