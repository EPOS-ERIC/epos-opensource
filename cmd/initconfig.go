package cmd

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/config"
	"github.com/spf13/cobra"
)

// initConfigCmd represents the init-config command
var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Create or replace the default user config file.",
	Long:  "Create or replace the default user config file. Writes epos-opensource.yaml to the standard config path for your platform. Use it to customize TUI settings and file or URL open commands.",
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
