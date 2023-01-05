package cmd

import (
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"mce.salesforce.com/sprinkler/database"
)

var withSample = false

// initializeCmd represents the initialize command
var initializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Initialize the database tables",
	Long:  `Initialize the database tables`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeDatabase()
	},
}

func initializeDatabase() {
	runDatabaseEffects(func(db *gorm.DB) {
		if db.Migrator().CreateTable(database.Tables...) != nil {
			panic("Failed creating tables")
		}
		if withSample {
			db.Create(&database.SampleWorkflows)
		}
	})
}

func init() {
	databaseCmd.AddCommand(initializeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initializeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	initializeCmd.Flags().BoolVar(&withSample, "with-sample", false, "With Sample Data")
}
