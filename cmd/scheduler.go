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
	Interval       time.Duration
	OrchardAddress string
}

func getSchedulerCmdOpt() SchedulerCmdOpt {
	return SchedulerCmdOpt{
		Interval:       viper.GetDuration("scheduler.interval"),
		OrchardAddress: viper.GetString("scheduler.orchardAddress"),
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
			Interval:    schedulerCmdOpt.Interval,
			MaxSize:     10,
			OrchardHost: schedulerCmdOpt.OrchardAddress,
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
		"orchard",
		"http://ws:8081",
		"address to orchard service",
	)
	viper.BindPFlag("scheduler.orchardAddress", schedulerCmd.Flags().Lookup("orchard"))
}
