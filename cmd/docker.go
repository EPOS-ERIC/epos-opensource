package cmd

import (
	"github.com/EPOS-ERIC/epos-opensource/cmd/docker"

	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Manage EPOS environments with Docker Compose.",
	Long:  "Manage EPOS environments with Docker Compose. Use these commands to deploy, update, list, populate, render, clean, and delete local environments.",
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
	dockerCmd.AddCommand(docker.GetCmd)
	dockerCmd.AddCommand(docker.ListCmd)
	dockerCmd.AddCommand(docker.CleanCmd)
	dockerCmd.AddCommand(docker.RenderCmd)
	rootCmd.AddCommand(dockerCmd)
}
