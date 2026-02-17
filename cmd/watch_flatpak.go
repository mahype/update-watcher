package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchFlatpakCmd = &cobra.Command{
	Use:   "flatpak",
	Short: "Add Flatpak application update watcher (Linux)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		watcher := config.WatcherConfig{
			Type:    "flatpak",
			Enabled: true,
			Options: map[string]interface{}{},
		}

		cfg.AddWatcher(watcher)

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Flatpak watcher added successfully.")
		return nil
	},
}

func init() {
	watchCmd.AddCommand(watchFlatpakCmd)
}
