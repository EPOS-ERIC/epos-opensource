package cmd

import (
	"epos-cli/cmd/k8s"

	"github.com/spf13/cobra"
)

var k8sCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Manage kubernetes environments",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	k8sCmd.AddCommand(k8s.DeployCmd)
	k8sCmd.AddCommand(k8s.DeleteCmd)
	// k8sCmd.AddCommand(docker.UpdateCmd)
	// k8sCmd.AddCommand(docker.PopulateCmd)
	rootCmd.AddCommand(k8sCmd)
}
