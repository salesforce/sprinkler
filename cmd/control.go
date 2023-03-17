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
	Address        string
	TrustedProxies []string
	APIKey         string
}

const (
	CtrlFlagAPIKey         string = "apiKey"
	CtrlConfigAPIKey              = "control.apiKey"
	CtrlFlagTrustedProxy          = "trustedProxy"
	CtrlConfigTrustedProxy        = "control.trustedProxies"
	CtrlFlagAddress               = "address"
	CtrlConfigAddress             = "control.address"
)

func getControlCmdOpt() ControlCmdOpt {
	return ControlCmdOpt{
		Address:        viper.GetString(CtrlConfigAddress),
		TrustedProxies: viper.GetStringSlice(CtrlConfigTrustedProxy),
		APIKey:         viper.GetString(CtrlConfigAPIKey),
	}
}

// controlCmd represents the control command
var controlCmd = &cobra.Command{
	Use:   "control",
	Short: "The control service for sprinkler",
	Long: `Control service provides interfaces to manage workflows that are registered
and run by sprinkler.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("control called")
		controlCmdOpt := getControlCmdOpt()
		ctrl := service.NewControl(
			controlCmdOpt.Address,
			controlCmdOpt.TrustedProxies,
			controlCmdOpt.APIKey,
		)
		ctrl.Run()
	},
}

func init() {
	serviceCmd.AddCommand(controlCmd)

	controlCmd.Flags().String(
		CtrlFlagAddress,
		":8080",
		"The address to listen to (e.g.: ':8080')",
	)
	viper.BindPFlag(CtrlConfigAddress, controlCmd.Flags().Lookup(CtrlFlagAddress))

	controlCmd.Flags().StringSlice(CtrlFlagTrustedProxy, []string{}, "trusted proxies")
	viper.BindPFlag(CtrlConfigTrustedProxy, controlCmd.Flags().Lookup(CtrlFlagTrustedProxy))

	controlCmd.Flags().String(CtrlFlagAPIKey, "", "api key")
	viper.BindPFlag(CtrlConfigAPIKey, controlCmd.Flags().Lookup(CtrlFlagAPIKey))

}
