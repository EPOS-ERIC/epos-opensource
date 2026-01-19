package docker

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore"
	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore/config"
	"github.com/EPOS-ERIC/epos-opensource/display"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy [env-name]",
	Short: "Create a new environment using Docker Compose.",
	Long:  "Deploy a new Docker Compose environment with the specified name.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		display.Debug("args: %v", args)
		display.Debug("configFilePath: %s", configFilePath)
		display.Debug("composeFilePath: %s", composeFilePath)
		display.Debug("path: %s", path)
		display.Debug("pullImages: %v", pullImages)

		name := ""
		if len(args) > 0 && args[0] != "" {
			name = args[0]
		}

		var cfg *config.EnvConfig
		var err error
		if configFilePath != "" {
			cfg, err = config.LoadConfigFromFile(configFilePath)
			if err != nil {
				display.Error("Failed to load config: %v", err)
				os.Exit(1)
			}
		}

		docker, err := dockercore.Deploy(dockercore.DeployOpts{
			EnvFile:     envFilePath,
			ComposeFile: composeFilePath,
			Path:        path,
			Name:        name,
			PullImages:  pullImages,
			Config:      cfg,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.URLs(docker.GuiUrl, docker.ApiUrl, fmt.Sprintf("epos-opensource docker deploy %s", name), docker.BackofficeUrl)
	},
}

func init() {
	DeployCmd.Flags().StringVar(&envFilePath, "env", "", "Path to the environment variables file (.env). If using a custom env file make sure to manually set the ports inside of it")
	DeployCmd.Flags().StringVar(&composeFilePath, "docker-compose", "", "Path to the Docker Compose file")
	DeployCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files")
	DeployCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Download Docker images before starting")
	DeployCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "Path to YAML configuration file")
}
