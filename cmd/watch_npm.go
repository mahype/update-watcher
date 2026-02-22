package cmd

import (
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("npm", "Add npm global package update watcher", "npm",
		nil,
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			return config.WatcherConfig{
				Type:    "npm",
				Enabled: true,
				Options: config.OptionsMap{},
			}, nil
		},
	)
}
