package k8s

import (
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"

	"github.com/spf13/cobra"
)

var ExportCmd = &cobra.Command{
	Use:   "export <path>",
	Short: "Write the default K8s config template.",
	Long:  "Write the default K8s config template. Exports a starter k8s-config.yaml file to the target directory.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		err := k8s.Export(k8s.ExportOpts{
			Path: path,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}
	},
}
