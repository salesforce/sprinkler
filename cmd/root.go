// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sprinkler",
	Short: "Workflow scheduler built for orchard",
	Long: `Sprinkler CLI provides all the necessary commands to setup the database, test
orchard servers and other services to run and manage workflows for orchard
orchestration server`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// persistent flags are globally available
	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default is $HOME/.sprinkler.yaml)",
	)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".sprinkler" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".sprinkler")
	}

	viper.AutomaticEnv()                                      // read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))    // replaces "." with "_" for environment variables
	viper.BindEnv("control.apiKeyEnabled", "API_KEY_ENABLED") // binds config key to environment variable
	viper.BindEnv("control.xfccEnabled", "XFCC_ENABLED")

	// If a config file is found, read it in, otherwise show error
	err := viper.ReadInConfig()
	if err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Fprintln(os.Stderr, "Error reading config file:", err)
	}
}
