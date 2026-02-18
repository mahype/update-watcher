package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/cron"
	"github.com/mahype/update-watcher/internal/rootcheck"
	"github.com/spf13/cobra"
)

var installCronCmd = &cobra.Command{
	Use:   "install-cron",
	Short: "Install a cron job for daily update checks",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("as-service-user")
		rootcheck.WarnOrReExec(force)

		cronExpr, _ := cmd.Flags().GetString("cron-expr")
		timeStr, _ := cmd.Flags().GetString("time")

		if cronExpr != "" {
			if err := cron.InstallWithExpr(cronExpr); err != nil {
				return err
			}
			fmt.Printf("Cron job installed with expression: %s\n", cronExpr)
		} else {
			if err := cron.Install(timeStr); err != nil {
				return err
			}
			fmt.Printf("Cron job installed for daily checks at %s\n", timeStr)
		}

		return nil
	},
}

func init() {
	installCronCmd.Flags().String("time", "07:00", "time of day for daily checks (HH:MM)")
	installCronCmd.Flags().String("cron-expr", "", "full cron expression (overrides --time)")
	rootCmd.AddCommand(installCronCmd)
}
