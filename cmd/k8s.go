package cmd

import (
	"github.com/epos-eu/epos-opensource/cmd/k8s"

	"github.com/spf13/cobra"
)

var k8sCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Manage Kubernetes environments.",
	Long:  "All Kubernetes commands use the current kubectl context configured on your system.",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	k8sCmd.AddCommand(k8s.DeployCmd)
	k8sCmd.AddCommand(k8s.DeleteCmd)
	k8sCmd.AddCommand(k8s.PopulateCmd)
	k8sCmd.AddCommand(k8s.ExportCmd)
	k8sCmd.AddCommand(k8s.UpdateCmd)
	k8sCmd.AddCommand(k8s.ListCmd)
	rootCmd.AddCommand(k8sCmd)
}
