// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mce.salesforce.com/sprinkler/service"
)

type OrchardCmd struct {
	Address string
}

var orchardCmdOpt OrchardCmd

var orchardAddress string

// orchardCmd represents the orchard command
var orchardCmd = &cobra.Command{
	Use:   "orchard",
	Short: "A fake orchard server",
	Long:  `Runs a fake orchard server for testing`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("orchard called")
		fo := service.NewFakeOrchard(orchardAddress)
		fo.Run()
	},
}

func init() {
	serviceCmd.AddCommand(orchardCmd)

	orchardCmd.Flags().StringVar(&orchardAddress, "address", ":8081", "The address to listen to (e.g.: ':8081'")
}
