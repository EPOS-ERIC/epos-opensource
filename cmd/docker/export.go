package docker

import (
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"

	"github.com/spf13/cobra"
)

var ExportCmd = &cobra.Command{
	Use:   "export [path]",
	Short: "Export the default Docker config to a directory.",
	Long:  "Export the default Docker configuration file (docker-config.yaml) to the specified directory.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		err := docker.Export(docker.ExportOpts{
			Path: path,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}
	},
}
