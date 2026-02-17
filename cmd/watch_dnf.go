package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchDnfCmd = &cobra.Command{
	Use:   "dnf",
	Short: "Add DNF package update watcher (Fedora/RHEL)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		securityOnly, _ := cmd.Flags().GetBool("security-only")
		noSudo, _ := cmd.Flags().GetBool("no-sudo")

		watcher := config.WatcherConfig{
			Type:    "dnf",
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

		fmt.Println("DNF watcher added successfully.")
		return nil
	},
}

func init() {
	watchDnfCmd.Flags().Bool("security-only", false, "only report security updates")
	watchDnfCmd.Flags().Bool("no-sudo", false, "do not use sudo for dnf operations")
	watchCmd.AddCommand(watchDnfCmd)
}
