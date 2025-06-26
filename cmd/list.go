package cmd

import (
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list installed environments",
	Run: func(cmd *cobra.Command, args []string) {
		dockerEnvs, err := db.GetEnvs("docker")
		if err != nil {
			return
		}

		kubeEnvs, err := db.GetEnvs("kubernetes")
		if err != nil {
			return
		}

		envs := append(dockerEnvs, kubeEnvs...)
		common.PrintEnvironmentList(envs, "all environments")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
