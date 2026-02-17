package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/cron"
	"github.com/spf13/cobra"
)

var uninstallCronCmd = &cobra.Command{
	Use:   "uninstall-cron",
	Short: "Remove the update-watcher cron job",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cron.Uninstall(); err != nil {
			return err
		}
		fmt.Println("Cron job removed successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCronCmd)
}
