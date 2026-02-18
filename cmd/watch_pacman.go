package cmd

import (
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("pacman", "Add Pacman package update watcher (Arch/Manjaro)", "Pacman",
		func(cmd *cobra.Command) {
			cmd.Flags().Bool("no-sudo", false, "do not use sudo for pacman sync operations")
		},
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			noSudo, _ := cmd.Flags().GetBool("no-sudo")
			return config.WatcherConfig{
				Type:    "pacman",
				Enabled: true,
				Options: config.OptionsMap{
					"use_sudo": !noSudo,
				},
			}, nil
		},
	)
}
