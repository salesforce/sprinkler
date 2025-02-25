// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"mce.salesforce.com/sprinkler/database"
	"mce.salesforce.com/sprinkler/service"
)

var controlAddress string

type ControlCmdOpt struct {
	Address         string
	TrustedProxies  []string
	APIKeyEnabled   bool
	APIKey          string
	XfccEnabled     bool
	XfccHeaderName  string
	XfccMustContain string
}

const (
	CtrlFlagAPIKeyEnabled     string = "apiKeyEnabled"
	CtrlConfigAPIKeyEnabled   string = "control.apiKeyEnabled"
	CtrlFlagAPIKey            string = "apiKey"
	CtrlConfigAPIKey          string = "control.apiKey"
	CtrlFlagXfccEnabled       string = "xfccEnabled"
	CtrlConfigXfccEnabled     string = "control.xfccEnabled"
	CtrlFlagXfccHeaderName    string = "xfccHeaderName"
	CtrlConfigXfccHeaderName  string = "control.xfccHeaderName"
	CtrlFlagXfccMustContain   string = "xfccMustContain"
	CtrlConfigXfccMustContain string = "control.xfccMustContain"
	CtrlFlagTrustedProxy      string = "trustedProxy"
	CtrlConfigTrustedProxy    string = "control.trustedProxies"
	CtrlFlagAddress           string = "address"
	CtrlConfigAddress         string = "control.address"
)

func getControlCmdOpt() ControlCmdOpt {
	return ControlCmdOpt{
		Address:         viper.GetString(CtrlConfigAddress),
		TrustedProxies:  viper.GetStringSlice(CtrlConfigTrustedProxy),
		APIKeyEnabled:   viper.GetBool(CtrlConfigAPIKeyEnabled),
		APIKey:          viper.GetString(CtrlConfigAPIKey),
		XfccEnabled:     viper.GetBool(CtrlConfigXfccEnabled),
		XfccHeaderName:  viper.GetString(CtrlConfigXfccHeaderName),
		XfccMustContain: viper.GetString(CtrlConfigXfccMustContain),
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
			database.GetInstance(),
			controlCmdOpt.Address,
			controlCmdOpt.TrustedProxies,
			controlCmdOpt.APIKeyEnabled,
			controlCmdOpt.APIKey,
			controlCmdOpt.XfccEnabled,
			controlCmdOpt.XfccHeaderName,
			controlCmdOpt.XfccMustContain,
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

	controlCmd.Flags().Bool(CtrlFlagAPIKeyEnabled, true, "api key enabled")
	viper.BindPFlag(CtrlConfigAPIKeyEnabled, controlCmd.Flags().Lookup(CtrlFlagAPIKeyEnabled))

	controlCmd.Flags().String(CtrlFlagAPIKey, "", "api key")
	viper.BindPFlag(CtrlConfigAPIKey, controlCmd.Flags().Lookup(CtrlFlagAPIKey))

	controlCmd.Flags().Bool(CtrlFlagXfccEnabled, true, "xfcc enabled")
	viper.BindPFlag(CtrlConfigXfccEnabled, controlCmd.Flags().Lookup(CtrlFlagXfccEnabled))

	controlCmd.Flags().String(CtrlFlagXfccHeaderName, "x-forwarded-client-cert", "xfcc header name")
	viper.BindPFlag(CtrlConfigXfccHeaderName, controlCmd.Flags().Lookup(CtrlFlagXfccHeaderName))

	controlCmd.Flags().String(CtrlFlagXfccMustContain, "", "xfcc must contain")
	viper.BindPFlag(CtrlConfigXfccMustContain, controlCmd.Flags().Lookup(CtrlFlagXfccMustContain))

}
