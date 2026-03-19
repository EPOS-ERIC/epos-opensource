package cmd

import (
	"github.com/EPOS-ERIC/epos-opensource/cmd/k8s"

	"github.com/spf13/cobra"
)

var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Manage EPOS environments on Kubernetes.",
	Long:  "Manage EPOS environments on Kubernetes. Most commands use the current kubectl context by default. The list command checks all contexts unless --context is set.",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	k8sCmd.AddCommand(k8s.DeployCmd)
	k8sCmd.AddCommand(k8s.DeleteCmd)
	k8sCmd.AddCommand(k8s.PopulateCmd)
	k8sCmd.AddCommand(k8s.ExportCmd)
	k8sCmd.AddCommand(k8s.GetCmd)
	k8sCmd.AddCommand(k8s.UpdateCmd)
	k8sCmd.AddCommand(k8s.ListCmd)
	k8sCmd.AddCommand(k8s.CleanCmd)
	k8sCmd.AddCommand(k8s.RenderCmd)
	rootCmd.AddCommand(k8sCmd)
}
