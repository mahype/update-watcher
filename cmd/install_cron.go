package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/cron"
	"github.com/mahype/update-watcher/internal/rootcheck"
	"github.com/spf13/cobra"
)

var installCronCmd = &cobra.Command{
	Use:   "install-cron",
	Short: "Install a cron job for scheduled tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("as-service-user")
		rootcheck.WarnOrReExec(force)

		cronExpr, _ := cmd.Flags().GetString("cron-expr")
		timeStr, _ := cmd.Flags().GetString("time")
		jobTypeStr, _ := cmd.Flags().GetString("type")

		jobType := cron.JobCheck
		switch jobTypeStr {
		case "self-update":
			jobType = cron.JobSelfUpdate
		case "check", "":
			jobType = cron.JobCheck
		default:
			return fmt.Errorf("unknown job type %q (use 'check' or 'self-update')", jobTypeStr)
		}

		if cronExpr != "" {
			if err := cron.InstallJobWithExpr(jobType, cronExpr); err != nil {
				return err
			}
			fmt.Printf("%s cron job installed with expression: %s\n", cron.JobTypeLabel(jobType), cronExpr)
		} else {
			if err := cron.InstallJob(jobType, timeStr); err != nil {
				return err
			}
			fmt.Printf("%s cron job installed for daily runs at %s\n", cron.JobTypeLabel(jobType), timeStr)
		}

		return nil
	},
}

func init() {
	installCronCmd.Flags().String("time", "07:00", "time of day for daily runs (HH:MM)")
	installCronCmd.Flags().String("cron-expr", "", "full cron expression (overrides --time)")
	installCronCmd.Flags().String("type", "check", "job type: check, self-update")
	rootCmd.AddCommand(installCronCmd)
}
