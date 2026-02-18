package cmd

import (
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("snap", "Add Snap package update watcher (Ubuntu/Linux)", "Snap",
		nil,
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			return config.WatcherConfig{
				Type:    "snap",
				Enabled: true,
				Options: config.OptionsMap{},
			}, nil
		},
	)
}
