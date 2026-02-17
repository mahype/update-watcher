package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchAptCmd = &cobra.Command{
	Use:   "apt",
	Short: "Add APT package update watcher",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		securityOnly, _ := cmd.Flags().GetBool("security-only")
		noSudo, _ := cmd.Flags().GetBool("no-sudo")

		watcher := config.WatcherConfig{
			Type:    "apt",
			Enabled: true,
			Options: map[string]interface{}{
				"use_sudo":      !noSudo,
				"security_only": securityOnly,
			},
		}

		cfg.AddWatcher(watcher)

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("APT watcher added successfully.")
		return nil
	},
}

func init() {
	watchAptCmd.Flags().Bool("security-only", false, "only report security updates")
	watchAptCmd.Flags().Bool("no-sudo", false, "do not use sudo for apt operations")
	watchCmd.AddCommand(watchAptCmd)
}
