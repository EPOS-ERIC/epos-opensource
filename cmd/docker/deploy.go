package docker

import (
	"github.com/epos-eu/epos-opensource/cmd/docker/internal"
	"github.com/epos-eu/epos-opensource/common"
	"fmt"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy [env-name]",
	Short: "Create a new environment using Docker Compose",
	Long:  "Deploys a new Docker Compose environment with the specified name.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		portalURL, gatewayURL, err := internal.Deploy(envFile, composeFile, path, name, pullImages)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
		common.PrintUrls(portalURL, gatewayURL, fmt.Sprintf("epos-opensource docker deploy %s", name))
	},
}

func init() {
	DeployCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Path to the environment variables file (.env)")
	DeployCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files")
	DeployCmd.Flags().StringVarP(&composeFile, "compose-file", "c", "", "Path to the Docker Compose file")
	DeployCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Download Docker images before starting")
}
