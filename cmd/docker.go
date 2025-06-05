package cmd

import (
	"epos-cli/cmd/docker"

	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker sub-command TODO",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	dockerCmd.AddCommand(docker.DeployCmd)
	dockerCmd.AddCommand(docker.DeleteCmd)
	dockerCmd.AddCommand(docker.RestartCmd)
	rootCmd.AddCommand(dockerCmd)
}
