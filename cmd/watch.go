package cmd

import (
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Add a watcher to the configuration",
	Long:  "Add a new update checker (apt, docker, or wordpress) to the configuration.",
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
