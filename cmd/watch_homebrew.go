package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchHomebrewCmd = &cobra.Command{
	Use:   "homebrew",
	Short: "Add Homebrew package update watcher (macOS/Linux)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		noCasks, _ := cmd.Flags().GetBool("no-casks")

		watcher := config.WatcherConfig{
			Type:    "homebrew",
			Enabled: true,
			Options: map[string]interface{}{
				"include_casks": !noCasks,
			},
		}

		cfg.AddWatcher(watcher)

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Homebrew watcher added successfully.")
		return nil
	},
}

func init() {
	watchHomebrewCmd.Flags().Bool("no-casks", false, "do not check cask updates")
	watchCmd.AddCommand(watchHomebrewCmd)
}
