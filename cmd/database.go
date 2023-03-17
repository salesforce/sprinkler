// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"mce.salesforce.com/sprinkler/database"
)

var dryRunOpt = false

// databaseCmd represents the database command
var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Utilities to manage and setup the database",
	Long:  `Manages database tables with the following commands:`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func runDatabaseEffects(effects func(*gorm.DB)) {
	db := database.GetInstance()
	effects(db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Info),
		DryRun: dryRunOpt,
	}))

	fmt.Println("Done!")
}

func init() {
	rootCmd.AddCommand(databaseCmd)

	// make this persistent flag available to all sub commands
	databaseCmd.PersistentFlags().BoolVar(&dryRunOpt, "dry-run", false, "Dry Run")

}
