package cmd

import (
	"fmt"
	"os"

	"github.com/mahype/update-watcher/cron"
	"github.com/mahype/update-watcher/internal/rootcheck"
	"github.com/mahype/update-watcher/internal/sudoers"
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
			removeSudoers()
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

// removeSudoers deletes /etc/sudoers.d/update-watcher when the full cron
// setup is torn down. Silently skips when the file does not exist or when we
// lack root privileges; a warning is printed if the removal itself fails.
func removeSudoers() {
	if _, err := os.Stat(sudoers.TargetPath); os.IsNotExist(err) {
		return
	}
	if !rootcheck.IsRoot() {
		fmt.Fprintln(os.Stderr, "WARN: not running as root — leaving "+sudoers.TargetPath+" in place.")
		fmt.Fprintln(os.Stderr, "      Re-run 'update-watcher uninstall-cron --all' as root to remove it.")
		return
	}
	if err := sudoers.Remove(); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: failed to remove sudoers file: %v\n", err)
		return
	}
	fmt.Println("Sudoers file removed: " + sudoers.TargetPath)
}

func init() {
	uninstallCronCmd.Flags().String("type", "check", "job type to remove: check, self-update")
	uninstallCronCmd.Flags().Bool("all", false, "remove all update-watcher cron jobs")
	rootCmd.AddCommand(uninstallCronCmd)
}
