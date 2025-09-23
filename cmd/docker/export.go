package docker

import (
	"os"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var ExportCmd = &cobra.Command{
	Use:   "export [path]",
	Short: "Export the default environment files to a directory.",
	Long:  "Export the default environment files: .env and docker-compose.yaml.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		err := dockercore.Export(dockercore.ExportOpts{
			Path: path,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}
	},
}
