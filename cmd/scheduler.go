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
var orchardAddress string

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
			Interval:    intervalOpt,
			MaxSize:     10,
			OrchardHost: orchardAddress,
		}
		scheduler.Start()
	},
}

func init() {
	serviceCmd.AddCommand(schedulerCmd)

	schedulerCmd.Flags().DurationVar(&intervalOpt, "interval", time.Minute, "scheduler check interval")
	schedulerCmd.Flags().StringVar(&orchardAddress, "orchard", "http://ws:8081", "address to orchard service")
}
