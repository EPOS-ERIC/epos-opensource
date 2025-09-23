package docker

import (
	"fmt"
	"os"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy [env-name]",
	Short: "Create a new environment using Docker Compose.",
	Long:  "Deploy a new Docker Compose environment with the specified name.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		docker, err := dockercore.Deploy(dockercore.DeployOpts{
			EnvFile:     envFile,
			ComposeFile: composeFile,
			Path:        path,
			Name:        name,
			PullImages:  pullImages,
			CustomHost:  host,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.Urls(docker.GuiUrl, docker.ApiUrl, docker.BackofficeUrl, fmt.Sprintf("epos-opensource docker deploy %s", name))
	},
}

func init() {
	DeployCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Path to the environment variables file (.env). If using a custom env file make sure to manually set the ports inside of it")
	DeployCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files")
	DeployCmd.Flags().StringVarP(&composeFile, "compose-file", "c", "", "Path to the Docker Compose file")
	DeployCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Download Docker images before starting")
	DeployCmd.Flags().StringVar(&host, "host", "", "Host (either IP or hostname) to use for exposing the environment")
}
