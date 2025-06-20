package docker

import (
	"epos-opensource/cmd/docker/internal"
	"epos-opensource/common"

	"github.com/spf13/cobra"
)

var ExportCmd = &cobra.Command{
	Use:   "export [path]",
	Short: "Export the default environment files to a directory",
	Long:  "Export the default environment files: .env and docker-compose.yaml.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		err := internal.Export(path)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
	},
}
