package cmd

import (
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("distro", "Add distribution release upgrade watcher", "Distro Release",
		func(cmd *cobra.Command) {
			cmd.Flags().Bool("lts-only", true, "only report LTS upgrades (Ubuntu only)")
		},
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			ltsOnly, _ := cmd.Flags().GetBool("lts-only")
			return config.WatcherConfig{
				Type:    "distro",
				Enabled: true,
				Options: config.OptionsMap{
					"lts_only": ltsOnly,
				},
			}, nil
		},
	)
}
