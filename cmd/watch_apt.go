package cmd

import (
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("apt", "Add APT package update watcher", "APT",
		func(cmd *cobra.Command) {
			cmd.Flags().Bool("security-only", false, "only report security updates")
			cmd.Flags().Bool("no-sudo", false, "do not use sudo for apt operations")
		},
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			securityOnly, _ := cmd.Flags().GetBool("security-only")
			noSudo, _ := cmd.Flags().GetBool("no-sudo")
			return config.WatcherConfig{
				Type:    "apt",
				Enabled: true,
				Options: config.OptionsMap{
					"use_sudo":      !noSudo,
					"security_only": securityOnly,
				},
			}, nil
		},
	)
}
