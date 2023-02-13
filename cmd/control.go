// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"mce.salesforce.com/sprinkler/service"
)

var controlAddress string

type ControlCmdOpt struct {
	Address string
}

func getControlCmdOpt() ControlCmdOpt {
	return ControlCmdOpt{
		Address: viper.GetString("control.address"),
	}
}

// controlCmd represents the control command
var controlCmd = &cobra.Command{
	Use:   "control",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("control called")
		controlCmdOpt := getControlCmdOpt()
		ctrl := service.NewControl(controlCmdOpt.Address)
		ctrl.Run()
	},
}

func init() {
	serviceCmd.AddCommand(controlCmd)

	controlCmd.Flags().String(
		"address",
		":8080",
		"The address to listen to (e.g.: ':8080')",
	)
	viper.BindPFlag("control.address", controlCmd.Flags().Lookup("address"))
}
