package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"mce.salesforce.com/sprinkler/service"
)

type CleanupCmdOpt struct {
	ScheduledWorkflowTimeout time.Duration
}

func getCleanupCmdOpt() CleanupCmdOpt {
	return CleanupCmdOpt{
		ScheduledWorkflowTimeout: viper.GetDuration("cleanup.scheduledWorkflowTimeout"),
	}
}

// schedulerCmd represents the scheduler command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Perform periodic database cleanup tasks",
	Long:  `Deletes records last updated earlier than the relevant timeout. Can be triggered on a cron.`,
	Run: func(cmd *cobra.Command, args []string) {
		cleanupCmdOpt := getCleanupCmdOpt()
		cleanup := &service.Cleanup{
			ScheduledWorkflowTimeout: cleanupCmdOpt.ScheduledWorkflowTimeout,
		}
		cleanup.Run()
	},
}

func init() {
	serviceCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().Duration(
		"scheduledWorkflowTimeout",
		time.Hour*24*30,
		"scheduled_workflow entries are considered expired if updated_at older than this duration",
	)
	viper.BindPFlag(
		"cleanup.scheduledWorkflowTimeout",
		schedulerCmd.Flags().Lookup("scheduledWorkflowTimeout"),
	)
}
