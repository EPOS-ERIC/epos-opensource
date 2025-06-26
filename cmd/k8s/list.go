package k8s

import (
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "list installed docker environments",
	Run: func(cmd *cobra.Command, args []string) {
		kubeEnvs, err := db.GetEnvs("kubernetes")
		if err != nil {
			return
		}

		common.PrintEnvironmentList(kubeEnvs, "all environments")
	},
}
