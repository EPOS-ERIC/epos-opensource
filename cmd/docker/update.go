package docker

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore"
	"github.com/EPOS-ERIC/epos-opensource/display"

	"github.com/spf13/cobra"
)

var (
	force bool
	reset bool
)

var UpdateCmd = &cobra.Command{
	Use:   "update [env-name]",
	Short: "Recreate an environment with new settings.",
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
			CustomHost:  host,
			Reset:       reset,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.Urls(d.GuiUrl, d.ApiUrl, d.BackofficeUrl, fmt.Sprintf("epos-opensource docker update %s", name))
	},
}

func init() {
	UpdateCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Path to the environment variables file (.env)")
	UpdateCmd.Flags().StringVarP(&composeFile, "compose-file", "c", "", "Path to the Docker Compose file")
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Remove the current containers before redeploying")
	UpdateCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Download Docker images before starting")
	UpdateCmd.Flags().StringVar(&host, "host", "", "Host (either IP or hostname) to use for exposing the environment")
	UpdateCmd.Flags().BoolVar(&reset, "reset", false, "Reset the environment config to the embedded defaults")
}
