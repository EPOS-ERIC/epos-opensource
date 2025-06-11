package docker

import (
	"epos-cli/cmd/docker/internal"
	"epos-cli/common"

	"github.com/spf13/cobra"
)

var force bool

var UpdateCmd = &cobra.Command{
	Use:   "update <name of an existing environment>",
	Short: "Recreate an environment with new settings",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		portalURL, gatewayURL, err := internal.Update(envFile, composeFile, path, name, force, pullImages)
		if err != nil {
			common.PrintError("%v", err)
			return
		}

		common.PrintUrls(portalURL, gatewayURL)
	},
}

func init() {
	UpdateCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files")
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Remove the current containers before redeploying")
	UpdateCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Download Docker images before starting")
}
