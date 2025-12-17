package cmd

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/config"
	"github.com/spf13/cobra"
)

// initConfigCmd represents the init-config command
var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Initialize the default configuration file",
	Long: `Creates the default configuration file at the standard location if it doesn't exist.
This allows users to customize settings like open commands for URLs, directories, and files.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.DefaultConfig()
		err := config.SaveConfig(cfg)
		if err != nil {
			fmt.Printf("Failed to create config file: %v\n", err)
			return
		}
		fmt.Printf("Config file created at %s\n", config.GetConfigPath())
	},
}

func init() {
	rootCmd.AddCommand(initConfigCmd)
}
