package cmd

import (
	"github.com/epos-eu/epos-opensource/tui/sample"
	"github.com/spf13/cobra"
)

// tuiSampleCmd runs the sample TUI.
// Demonstrates screen-based architecture: registers screens,
// starts app, and handles navigation without import cycles.
var tuiSampleCmd = &cobra.Command{
	Use:   "tui-sample",
	Short: "Run the sample TUI for testing screen-based structure",
	Long:  `This command runs a sample TUI implementation using the screen-based architecture.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app := sample.NewApp(nil)

		// Register screens: App holds interfaces, screens implement them
		app.RegisterScreen(&sample.HomeScreen{})
		app.RegisterScreen(&sample.DeployScreen{})

		return app.Run()
	},
}

func init() {
	rootCmd.AddCommand(tuiSampleCmd)
}
