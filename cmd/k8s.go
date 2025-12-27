package cmd

import (
	"github.com/epos-eu/epos-opensource/cmd/k8s"

	"github.com/spf13/cobra"
)

var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Manage K8s environments.",
	Long:  "All K8s commands use the current kubectl context configured on your system.",
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
	k8sCmd.AddCommand(k8s.CleanCmd)
	rootCmd.AddCommand(k8sCmd)
}
