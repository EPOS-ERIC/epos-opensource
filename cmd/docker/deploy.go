package docker

import (
	"epos-cli/cmd/docker/internal"
	"epos-cli/common"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy <name to give to the environment>",
	Short: "docker deploy cmd TODO",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := internal.Deploy(envFile, composeFile, path, name, pullImages)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
	},
}

func init() {
	DeployCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Environment variable file, use default if not provided")
	DeployCmd.Flags().StringVarP(&path, "path", "p", "", "Custom path for the creation of the dir with the env and compose files")
	DeployCmd.Flags().StringVarP(&composeFile, "compose-file", "c", "", "Docker compose file, use default if not provided")
	DeployCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "If set the images for the environment will be pulled before deploying")
}
