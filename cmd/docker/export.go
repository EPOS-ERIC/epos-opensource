package docker

import (
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"

	"github.com/spf13/cobra"
)

var ExportCmd = &cobra.Command{
	Use:   "export <path>",
	Short: "Write the default Docker config template.",
	Long:  "Write the default Docker config template. Exports a starter docker-config.yaml file to the target directory.",
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
