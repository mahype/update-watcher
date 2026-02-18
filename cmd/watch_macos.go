package cmd

import (
	"fmt"
	"runtime"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

func init() {
	addWatchCommand("macos", "Add macOS software update watcher", "macOS",
		func(cmd *cobra.Command) {
			cmd.Flags().Bool("security-only", false, "only report security updates")
		},
		func(cmd *cobra.Command) (config.WatcherConfig, error) {
			if runtime.GOOS != "darwin" {
				fmt.Println("Warning: macOS watcher is intended for macOS systems.")
			}
			securityOnly, _ := cmd.Flags().GetBool("security-only")
			return config.WatcherConfig{
				Type:    "macos",
				Enabled: true,
				Options: config.OptionsMap{
					"security_only": securityOnly,
				},
			}, nil
		},
	)
}
