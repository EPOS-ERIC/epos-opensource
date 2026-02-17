package k8s

import (
	"os"

	"github.com/EPOS-ERIC/epos-opensource/cmd/k8s/k8score"
	"github.com/EPOS-ERIC/epos-opensource/display"

	"github.com/spf13/cobra"
)

var ExportCmd = &cobra.Command{
	Use:   "export [path]",
	Short: "Export default environment files and manifests.",
	Long:  "Copies the default .env file and all embedded K8s manifest files to the specified directory for manual inspection or customization.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		err := k8score.Export(k8score.ExportOpts{
			Path: path,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}
	},
}
