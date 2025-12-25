package k8s

import (
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed K8s environments.",
	Run: func(cmd *cobra.Command, args []string) {
		kubeEnvs, err := db.GetAllK8s()
		if err != nil {
			return
		}

		display.K8sList(kubeEnvs, "Installed K8s environments")
	},
}
