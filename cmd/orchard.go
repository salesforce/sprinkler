/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
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
