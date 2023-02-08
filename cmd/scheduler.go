// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"mce.salesforce.com/sprinkler/service"
)

var intervalOpt time.Duration

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
		scheduler := &service.Scheduler{
			Interval: intervalOpt,
			MaxSize:  10,
		}
		scheduler.Start()
	},
}

func init() {
	serviceCmd.AddCommand(schedulerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// schedulerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// schedulerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	schedulerCmd.Flags().DurationVar(&intervalOpt, "interval", time.Minute, "scheduler check interval")
}
