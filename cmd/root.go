package cmd

import (
	"fmt"
	"os"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "update-watcher",
	Short: "Monitor servers for available updates",
	Long:  "A modular tool that checks for system and application updates (APT, DNF, Pacman, Zypper, APK, Docker, WordPress, macOS) and sends notifications.",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: "+config.DefaultConfigDir()+"/config.yaml)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress terminal output")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable debug logging")

	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		for _, p := range config.ConfigSearchPaths() {
			viper.AddConfigPath(p)
		}
	}

	viper.SetEnvPrefix("UPDATE_WATCHER")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config: %s\n", err)
		}
	}
}
