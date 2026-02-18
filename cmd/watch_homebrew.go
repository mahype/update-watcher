package cmd

import (
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("homebrew", "Add Homebrew package update watcher (macOS/Linux)", "Homebrew",
		func(cmd *cobra.Command) {
			cmd.Flags().Bool("no-casks", false, "do not check cask updates")
		},
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			noCasks, _ := cmd.Flags().GetBool("no-casks")
			return config.WatcherConfig{
				Type:    "homebrew",
				Enabled: true,
				Options: config.OptionsMap{
					"include_casks": !noCasks,
				},
			}, nil
		},
	)
}
