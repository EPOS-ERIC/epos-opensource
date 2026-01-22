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
		display.Debug("path: %s", path)
		display.Debug("pullImages: %v", pullImages)

		var cfg *config.EnvConfig
		var err error
		if configFilePath == "" {
			cfg = config.GetDefaultConfig()
		} else {
			cfg, err = config.LoadConfig(configFilePath)
			if err != nil {
				display.Error("Failed to load config: %v", err)
				os.Exit(1)
			}
		}

		if len(args) > 0 && args[0] != "" {
			cfg.Name = args[0]
			if configFilePath != "" {
				display.Warn("Using environment name from command line: %s", cfg.Name)
			}
		}

		docker, err := dockercore.Deploy(dockercore.DeployOpts{
			Path:       path,
			PullImages: pullImages,
			Config:     cfg,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.URLs(docker.GuiUrl, docker.ApiUrl, fmt.Sprintf("epos-opensource docker deploy %s", docker.Name), docker.BackofficeUrl)
	},
}

func init() {
	DeployCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files")
	DeployCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Download Docker images before starting")
	DeployCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "Path to YAML configuration file")
}
