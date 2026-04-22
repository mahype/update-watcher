package cmd

import (
	"fmt"
	"os"

	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/cron"
	"github.com/mahype/update-watcher/internal/rootcheck"
	"github.com/mahype/update-watcher/internal/sudoers"
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

		return installSudoers()
	},
}

// installSudoers generates /etc/sudoers.d/update-watcher from the current
// config. Failure to load the config or to write the file is reported but
// does not fail the install-cron command — the cron job is already in place.
func installSudoers() error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nWARN: skipping sudoers generation — config not loadable: %v\n", err)
		return nil
	}

	rules, warnings, err := sudoers.Build(cfg)
	if err != nil {
		return fmt.Errorf("sudoers: %w", err)
	}
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "WARN:", w)
	}
	if len(rules) == 0 {
		return nil
	}

	if !rootcheck.IsRoot() {
		fmt.Fprintln(os.Stderr, "\nWARN: not running as root — skipping sudoers file generation.")
		fmt.Fprintln(os.Stderr, "      Re-run 'update-watcher install-cron' as root to generate "+sudoers.TargetPath+".")
		return nil
	}

	if err := sudoers.Write(rules); err != nil {
		return fmt.Errorf("sudoers: %w", err)
	}

	fmt.Printf("\nSudoers rules written to %s:\n", sudoers.TargetPath)
	serviceUser := rootcheck.ServiceUserName()
	for _, r := range rules {
		fmt.Println("  " + sudoers.FormatRule(serviceUser, r))
	}
	return nil
}

func init() {
	installCronCmd.Flags().String("time", "07:00", "time of day for daily runs (HH:MM)")
	installCronCmd.Flags().String("cron-expr", "", "full cron expression (overrides --time)")
	installCronCmd.Flags().String("type", "check", "job type: check, self-update")
	rootCmd.AddCommand(installCronCmd)
}
