package cmd

import (
	"fmt"
	"strings"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchDockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Add Docker container update watcher",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		containers, _ := cmd.Flags().GetString("containers")
		excludeStr, _ := cmd.Flags().GetStringSlice("exclude")

		options := map[string]interface{}{
			"containers": containers,
		}
		if len(excludeStr) > 0 {
			exclude := make([]interface{}, len(excludeStr))
			for i, s := range excludeStr {
				exclude[i] = strings.TrimSpace(s)
			}
			options["exclude"] = exclude
		}

		watcher := config.WatcherConfig{
			Type:    "docker",
			Enabled: true,
			Options: options,
		}

		cfg.AddWatcher(watcher)

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Docker watcher added successfully.")
		return nil
	},
}

func init() {
	watchDockerCmd.Flags().String("containers", "all", "comma-separated container names, or \"all\"")
	watchDockerCmd.Flags().StringSlice("exclude", nil, "container names to exclude")
	watchCmd.AddCommand(watchDockerCmd)
}
