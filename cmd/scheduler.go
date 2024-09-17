// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"os"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"mce.salesforce.com/sprinkler/service"
)

type SchedulerCmdOpt struct {
	Interval          time.Duration
	OrchardAddress    string
	OrchardAPIKeyName string
	OrchardAPIKey     string
}

func getSchedulerCmdOpt() SchedulerCmdOpt {
	schedulerInterval := viper.GetString("scheduler.interval")
	if schedulerInterval == "" {
		schedulerInterval = os.Getenv("SCHEDULER_INTERVAL")
	}
	orchardAddress := viper.GetString("scheduler.orchard.address")
	if orchardAddress == "" {
		orchardAddress = os.Getenv("SCHEDULER_ORCHARD_ADDRESS")
	}
	orchardAPIKeyName := viper.GetString("scheduler.orchard.apiKeyName")
	if orchardAPIKeyName == "" {
		orchardAPIKeyName = os.Getenv("SCHEDULER_ORCHARD_API_KEY_NAME")
	}
	return SchedulerCmdOpt{
		Interval:          cast.ToDuration(schedulerInterval),
		OrchardAddress:    orchardAddress,
		OrchardAPIKeyName: orchardAPIKeyName,
		OrchardAPIKey:     viper.GetString("scheduler.orchard.apiKey"),
	}
}

// schedulerCmd represents the scheduler command
var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		schedulerCmdOpt := getSchedulerCmdOpt()
		scheduler := &service.Scheduler{
			Interval:          schedulerCmdOpt.Interval,
			MaxSize:           10,
			OrchardHost:       schedulerCmdOpt.OrchardAddress,
			OrchardAPIKeyName: schedulerCmdOpt.OrchardAPIKeyName,
			OrchardAPIKey:     schedulerCmdOpt.OrchardAPIKey,
		}
		scheduler.Start()
	},
}

func init() {
	serviceCmd.AddCommand(schedulerCmd)

	schedulerCmd.Flags().Duration(
		"interval",
		time.Minute,
		"scheduler check interval",
	)
	viper.BindPFlag("scheduler.interval", schedulerCmd.Flags().Lookup("interval"))

	schedulerCmd.Flags().String(
		"orchardAddress",
		"http://ws:8081",
		"address to orchard service",
	)
	viper.BindPFlag("scheduler.orchard.address", schedulerCmd.Flags().Lookup("orchard"))

	schedulerCmd.Flags().String(
		"orchardAPIKeyName",
		"",
		"api key name to orchard service",
	)
	viper.BindPFlag("scheduler.orchard.apiKeyName", schedulerCmd.Flags().Lookup("orchard"))

	schedulerCmd.Flags().String(
		"orchardAPIKey",
		"",
		"api key to orchard service",
	)
	viper.BindPFlag("scheduler.orchard.apiKey", schedulerCmd.Flags().Lookup("orchard"))
}
