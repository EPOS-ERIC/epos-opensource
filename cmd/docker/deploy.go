package docker

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy [env-name]",
	Short: "Create a new environment using Docker Compose.",
	Long:  "Deploy a new Docker Compose environment with the specified name.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

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

		cfg.Name = name

		env, err := docker.Deploy(docker.DeployOpts{
			Path:       path,
			PullImages: pullImages,
			Config:     cfg,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.URLs(env.GuiUrl, env.ApiUrl, fmt.Sprintf("epos-opensource docker deploy %s", env.Name), env.BackofficeUrl)
	},
}

func init() {
	DeployCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files")
	DeployCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Download Docker images before starting")
	DeployCmd.Flags().StringVar(&configFilePath, "config", "", "Path to YAML configuration file")
}
