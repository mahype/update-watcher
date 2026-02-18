package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := config.Validate(cfg); err != nil {
			return err
		}

		if warnings := config.WarnPlaintextSecrets(cfg); len(warnings) > 0 {
			for _, w := range warnings {
				fmt.Printf("\u26a0  %s\n", w)
			}
			fmt.Println()
		}

		fmt.Printf("Configuration is valid (%s)\n", config.ConfigPath())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
