package k8s

import (
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "list installed kubernetes environments",
	Run: func(cmd *cobra.Command, args []string) {
		kubeEnvs, err := db.GetAllKubernetes()
		if err != nil {
			return
		}

		display.KubernetesList(kubeEnvs, "installed kubernetes environments")
	},
}
