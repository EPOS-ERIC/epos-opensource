package docker

import (
	"epos-opensource/cmd/docker/internal"
	"epos-opensource/common"

	"github.com/spf13/cobra"
)

var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the default environment files",
	Long:  "Export the default environment files: .env and docker-compose.yaml.",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := internal.Export(path)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
	},
}

func init() {
	ExportCmd.Flags().StringVarP(&path, "path", "p", "", "Location to export the files to. CWD by default")
}
