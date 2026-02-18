package cmd

import (
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("openclaw", "Add OpenClaw update watcher", "OpenClaw",
		func(cmd *cobra.Command) {
			cmd.Flags().String("channel", "", "update channel (stable, beta, dev)")
		},
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			channel, _ := cmd.Flags().GetString("channel")
			opts := config.OptionsMap{}
			if channel != "" {
				opts["channel"] = channel
			}
			return config.WatcherConfig{
				Type:    "openclaw",
				Enabled: true,
				Options: opts,
			}, nil
		},
	)
}
