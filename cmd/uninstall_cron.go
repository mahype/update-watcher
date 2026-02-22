package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/cron"
	"github.com/spf13/cobra"
)

var uninstallCronCmd = &cobra.Command{
	Use:   "uninstall-cron",
	Short: "Remove update-watcher cron jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		jobTypeStr, _ := cmd.Flags().GetString("type")

		if all {
			if err := cron.UninstallAll(); err != nil {
				return err
			}
			fmt.Println("All update-watcher cron jobs removed.")
			return nil
		}

		jobType := cron.JobCheck
		switch jobTypeStr {
		case "self-update":
			jobType = cron.JobSelfUpdate
		case "check", "":
			jobType = cron.JobCheck
		default:
			return fmt.Errorf("unknown job type %q (use 'check' or 'self-update')", jobTypeStr)
		}

		if err := cron.UninstallJob(jobType); err != nil {
			return err
		}
		fmt.Printf("%s cron job removed successfully.\n", cron.JobTypeLabel(jobType))
		return nil
	},
}

func init() {
	uninstallCronCmd.Flags().String("type", "check", "job type to remove: check, self-update")
	uninstallCronCmd.Flags().Bool("all", false, "remove all update-watcher cron jobs")
	rootCmd.AddCommand(uninstallCronCmd)
}
