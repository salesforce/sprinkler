// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"mce.salesforce.com/sprinkler/service"
)

type SchedulerCmdOpt struct {
	Interval                 time.Duration
	LockTimeout              time.Duration
	OrchardAddress           string
	OrchardAPIKeyName        string
	OrchardAPIKey            string
	ScheduledWorkflowTimeout time.Duration
}

func getSchedulerCmdOpt() SchedulerCmdOpt {
	return SchedulerCmdOpt{
		Interval:                 viper.GetDuration("scheduler.interval"),
		LockTimeout:              viper.GetDuration("scheduler.lockTimeout"),
		OrchardAddress:           viper.GetString("scheduler.orchard.address"),
		OrchardAPIKeyName:        viper.GetString("scheduler.orchard.apiKeyName"),
		OrchardAPIKey:            viper.GetString("scheduler.orchard.apiKey"),
		ScheduledWorkflowTimeout: viper.GetDuration("scheduler.scheduledWorkflowTimeout"),
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
			LockTimeout:       schedulerCmdOpt.LockTimeout,
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

	schedulerCmd.Flags().Duration(
		"lockTimeout",
		time.Hour,
		"Workflow schedule and activation lock TTL",
	)
	viper.BindPFlag("scheduler.lockTimeout", schedulerCmd.Flags().Lookup("lockTimeout"))

	schedulerCmd.Flags().Duration(
		"scheduledWorkflowTimeout",
		time.Hour*24*30,
		"scheduled_workflow entries are considered expired if updated_at older than this duration",
	)
	viper.BindPFlag(
		"scheduler.scheduledWorkflowTimeout",
		schedulerCmd.Flags().Lookup("scheduledWorkflowTimeout"),
	)
}
