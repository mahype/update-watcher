package cmd

import (
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("apk", "Add APK package update watcher (Alpine Linux)", "APK",
		func(cmd *cobra.Command) {
			cmd.Flags().Bool("no-sudo", false, "do not use sudo for apk operations")
		},
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			noSudo, _ := cmd.Flags().GetBool("no-sudo")
			return config.WatcherConfig{
				Type:    "apk",
				Enabled: true,
				Options: config.OptionsMap{
					"use_sudo": !noSudo,
				},
			}, nil
		},
	)
}
