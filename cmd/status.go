package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/cron"
	"github.com/mahype/update-watcher/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current configuration and status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")

		cronInstalled, cronSchedule := cron.IsInstalled()

		switch format {
		case "json":
			data := map[string]interface{}{
				"config":         cfg,
				"cron_installed": cronInstalled,
				"cron_schedule":  cronSchedule,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(data)
		default:
			output.PrintStatus(cfg, cronInstalled, cronSchedule)
		}

		return nil
	},
}

func init() {
	statusCmd.Flags().StringP("format", "f", "table", "output format: table, json")
	rootCmd.AddCommand(statusCmd)
}
