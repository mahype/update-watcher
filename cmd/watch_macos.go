package cmd

import (
	"fmt"
	"runtime"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchMacosCmd = &cobra.Command{
	Use:   "macos",
	Short: "Add macOS software update watcher",
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS != "darwin" {
			fmt.Println("Warning: macOS watcher is intended for macOS systems.")
		}

		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		securityOnly, _ := cmd.Flags().GetBool("security-only")

		watcher := config.WatcherConfig{
			Type:    "macos",
			Enabled: true,
			Options: map[string]interface{}{
				"security_only": securityOnly,
			},
		}

		cfg.AddWatcher(watcher)

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("macOS watcher added successfully.")
		return nil
	},
}

func init() {
	watchMacosCmd.Flags().Bool("security-only", false, "only report security updates")
	watchCmd.AddCommand(watchMacosCmd)
}
