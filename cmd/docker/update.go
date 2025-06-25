package docker

import (
	"fmt"
	"github.com/epos-eu/epos-opensource/cmd/docker/internal"
	"github.com/epos-eu/epos-opensource/common"

	"github.com/spf13/cobra"
)

var force bool

var UpdateCmd = &cobra.Command{
	Use:   "update [env-name]",
	Short: "Recreate an environment with new settings",
	Long:  "Re-deploy an existing Docker Compose environment after modifying its configuration.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		portalURL, gatewayURL, err := internal.Update(envFile, composeFile, path, name, force, pullImages)
		if err != nil {
			common.PrintError("%v", err)
			return
		}

		common.PrintUrls(portalURL, gatewayURL, fmt.Sprintf("epos-opensource docker update %s", name))
	},
}

func init() {
	UpdateCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Path to the environment variables file (.env)")
	UpdateCmd.Flags().StringVarP(&composeFile, "compose-file", "c", "", "Path to the Docker Compose file")
	UpdateCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files")
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Remove the current containers before redeploying")
	UpdateCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Download Docker images before starting")
}
