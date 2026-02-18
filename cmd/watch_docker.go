package cmd

import (
	"strings"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("docker", "Add Docker container update watcher", "Docker",
		func(cmd *cobra.Command) {
			cmd.Flags().String("containers", "all", "comma-separated container names, or \"all\"")
			cmd.Flags().StringSlice("exclude", nil, "container names to exclude")
		},
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			containers, _ := cmd.Flags().GetString("containers")
			excludeStr, _ := cmd.Flags().GetStringSlice("exclude")

			options := config.OptionsMap{
				"containers": containers,
			}
			if len(excludeStr) > 0 {
				exclude := make([]interface{}, len(excludeStr))
				for i, s := range excludeStr {
					exclude[i] = strings.TrimSpace(s)
				}
				options["exclude"] = exclude
			}

			return config.WatcherConfig{
				Type:    "docker",
				Enabled: true,
				Options: options,
			}, nil
		},
	)
}
