package cmd

import (
	"github.com/epos-eu/epos-opensource/cmd/docker"

	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Manage local Docker Compose environments.",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	dockerCmd.AddCommand(docker.DeployCmd)
	dockerCmd.AddCommand(docker.DeleteCmd)
	dockerCmd.AddCommand(docker.UpdateCmd)
	dockerCmd.AddCommand(docker.PopulateCmd)
	dockerCmd.AddCommand(docker.ExportCmd)
	dockerCmd.AddCommand(docker.ListCmd)
	rootCmd.AddCommand(dockerCmd)
}
