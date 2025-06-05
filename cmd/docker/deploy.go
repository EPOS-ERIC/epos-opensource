package docker

import (
	"epos-cli/cmd/docker/internal"
	"epos-cli/common"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy <name>",
	Short: "docker deploy cmd TODO",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := internal.Deploy(envFile, composeFile, path, name)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
	},
}

func init() {
	addCommonFlags(DeployCmd)
}
