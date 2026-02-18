package cmd

import (
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("flatpak", "Add Flatpak application update watcher (Linux)", "Flatpak",
		nil,
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			return config.WatcherConfig{
				Type:    "flatpak",
				Enabled: true,
				Options: config.OptionsMap{},
			}, nil
		},
	)
}
