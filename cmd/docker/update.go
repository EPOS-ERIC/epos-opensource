package docker

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/display"

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

		d, err := dockercore.Update(dockercore.UpdateOpts{
			EnvFile:     envFile,
			ComposeFile: composeFile,
			Name:        name,
			PullImages:  pullImages,
			Force:       force,
		})
		if err != nil {
			display.Error("%v", err)
			return
		}

		display.Urls(d.GuiUrl, d.ApiUrl, d.BackofficeUrl, fmt.Sprintf("epos-opensource docker update %s", name))
	},
}

func init() {
	UpdateCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Path to the environment variables file (.env)")
	UpdateCmd.Flags().StringVarP(&composeFile, "compose-file", "c", "", "Path to the Docker Compose file")
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Remove the current containers before redeploying")
	UpdateCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Download Docker images before starting")
}
