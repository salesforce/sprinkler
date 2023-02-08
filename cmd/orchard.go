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

// orchardCmd represents the orchard command
var orchardCmd = &cobra.Command{
	Use:   "orchard",
	Short: "A fake orchard server",
	Long:  `Runs a fake orchard server for testing`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("orchard called")
		fo := service.NewFakeOrchard()
		fo.Run()
	},
}

func init() {
	serviceCmd.AddCommand(orchardCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// orchardCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// orchardCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
