// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"mce.salesforce.com/sprinkler/database"
)

var dryRunOpt = false

type DatabaseCmdOpt struct {
	Host     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

const (
	DBFlagHost       string = "host"
	DBConfigHost            = "db.host"
	DBFlagUser              = "user"
	DBConfigUser            = "db.user"
	DBFlagPassword          = "password"
	DBConfigPassword        = "db.password"
	DBFlagDBName            = "dbname"
	DBConfigDBName          = "db.dbname"
	DBFlagSSLMode           = "sslmode"
	DBConfigSSLMode         = "db.sslmode"
)

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
	databaseCmd.PersistentFlags().BoolVar(
		&dryRunOpt,
		"dry-run",
		false,
		"Dry Run",
	)

	databaseCmd.PersistentFlags().String(DBFlagHost, "db", "database host")
	viper.BindPFlag(
		DBConfigHost,
		databaseCmd.PersistentFlags().Lookup(DBFlagHost),
	)

	databaseCmd.PersistentFlags().String(DBFlagUser, "postgres", "database username")
	viper.BindPFlag(
		DBConfigUser,
		databaseCmd.PersistentFlags().Lookup(DBFlagUser),
	)

	databaseCmd.PersistentFlags().String(DBFlagPassword, "sprinkler", "database password")
	viper.BindPFlag(
		DBConfigPassword,
		databaseCmd.PersistentFlags().Lookup(DBFlagPassword),
	)

	databaseCmd.PersistentFlags().String(DBFlagDBName, "postgres", "database name")
	viper.BindPFlag(
		DBConfigDBName,
		databaseCmd.PersistentFlags().Lookup(DBFlagDBName),
	)

	databaseCmd.PersistentFlags().String(DBFlagSSLMode, "disable", "database name")
	viper.BindPFlag(
		DBConfigSSLMode,
		databaseCmd.PersistentFlags().Lookup(DBFlagSSLMode),
	)
}
