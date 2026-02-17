package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchSnapCmd = &cobra.Command{
	Use:   "snap",
	Short: "Add Snap package update watcher (Ubuntu/Linux)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		watcher := config.WatcherConfig{
			Type:    "snap",
			Enabled: true,
			Options: map[string]interface{}{},
		}

		cfg.AddWatcher(watcher)

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Snap watcher added successfully.")
		return nil
	},
}

func init() {
	watchCmd.AddCommand(watchSnapCmd)
}
