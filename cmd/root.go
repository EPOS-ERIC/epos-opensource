package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "epos-cli",
	Short: "Manage ePOS environments and utilities",
	Long: `epos-cli provides commands for working with local Docker environments
used by the ePOS project. Use the "docker" group of commands to deploy, update
or delete an environment.`,

	// If no subcommand is provided, show help.
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
