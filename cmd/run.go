package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/output"
	"github.com/mahype/update-watcher/runner"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all configured update checks",
	Long:  "Execute all enabled watchers and send notifications.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		only, _ := cmd.Flags().GetString("only")
		format, _ := cmd.Flags().GetString("format")
		quiet, _ := cmd.Flags().GetBool("quiet")

		r := runner.New(cfg,
			runner.WithDryRun(dryRun),
			runner.WithOnly(only),
		)

		result, err := r.Run()
		if err != nil {
			return err
		}

		// Output results
		if !quiet {
			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(result)
			default:
				output.PrintResults(result.Results, result.Errors)
			}
		}

		// Exit code based on results
		if len(result.Errors) > 0 && result.TotalUpdates == 0 {
			os.Exit(3)
		}
		if len(result.Errors) > 0 {
			os.Exit(2)
		}
		if result.TotalUpdates > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	runCmd.Flags().BoolP("dry-run", "n", false, "run checks but skip notifications")
	runCmd.Flags().StringP("format", "f", "text", "output format: text, json")
	runCmd.Flags().String("only", "", "run only a specific checker: apt, docker, wordpress")
	runCmd.Flags().Bool("notify", true, "enable/disable notifications")
	rootCmd.AddCommand(runCmd)
}
