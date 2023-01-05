package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"mce.salesforce.com/sprinkler/database"
)

var dryRunOpt = false

// databaseCmd represents the database command
var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "A brief description of your command",
	Long:  `Manages database tables with the following commands:`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func runDatabaseEffects(effects func(*gorm.DB)) {
	db := database.GetInstance()
	effects(db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Info),
		DryRun: dryRunOpt,
	}))

	fmt.Println("Done!")
}

func init() {
	rootCmd.AddCommand(databaseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	databaseCmd.PersistentFlags().BoolVar(&dryRunOpt, "dry-run", false, "Dry Run")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// databaseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
