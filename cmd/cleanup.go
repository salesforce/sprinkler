package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"mce.salesforce.com/sprinkler/service"
)

type CleanupCmdOpt struct {
	ScheduledWorkflowTimeout      time.Duration
	WorkflowActivationLockTimeout time.Duration
	WorkflowSchedulerLockTimeout  time.Duration
}

func getCleanupCmdOpt() CleanupCmdOpt {
	return CleanupCmdOpt{
		ScheduledWorkflowTimeout:      viper.GetDuration("cleanup.scheduledWorkflowTimeout"),
		WorkflowActivationLockTimeout: viper.GetDuration("cleanup.workflowActivationLock"),
		WorkflowSchedulerLockTimeout:  viper.GetDuration("cleanup.workflowSchedulerLock"),
	}
}

// represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Perform periodic database cleanup tasks",
	Long:  `Deletes records last updated earlier than the relevant timeout. Can be triggered on a cron.`,
	Run: func(cmd *cobra.Command, args []string) {
		cleanupCmdOpt := getCleanupCmdOpt()
		cleanup := &service.Cleanup{
			ScheduledWorkflowTimeout:      cleanupCmdOpt.ScheduledWorkflowTimeout,
			WorkflowActivationLockTimeout: cleanupCmdOpt.WorkflowActivationLockTimeout,
			WorkflowSchedulerLockTimeout:  cleanupCmdOpt.WorkflowSchedulerLockTimeout,
		}
		cleanup.Run()
	},
}

func init() {
	serviceCmd.AddCommand(cleanupCmd)

	cleanupCmd.Flags().Duration(
		"scheduledWorkflow",
		time.Hour*24*30,
		"scheduled_workflow entries are considered expired if updated_at older than this duration",
	)
	viper.BindPFlag("cleanup.scheduledWorkflow", cleanupCmd.Flags().Lookup("scheduledWorkflow"))

	cleanupCmd.Flags().Duration(
		"workflowActivationLock",
		time.Hour,
		"Workflow activation lock TTL",
	)
	viper.BindPFlag("cleanup.workflowActivationLock", cleanupCmd.Flags().Lookup("workflowActivationLock"))

	cleanupCmd.Flags().Duration(
		"workflowSchedulerLock",
		time.Hour,
		"Workflow scheduleer lock TTL",
	)
	viper.BindPFlag("cleanup.workflowSchedulerLock", cleanupCmd.Flags().Lookup("workflowSchedulerLock"))
}
