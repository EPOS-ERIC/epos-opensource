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
	Use:   "deploy <env-name>",
	Short: "Deploy a new environment.",
	Long:  "Deploy a new environment. Starts a new local Docker Compose environment with the given name. Uses the default configuration unless --config is set.",
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
			PullImages: pullImages,
			Config:     cfg,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		urls, err := env.BuildEnvURLs()
		if err != nil {
			display.Error("failed to build environment URLs: %v", err)
			os.Exit(1)
		}

		display.URLs(urls.GUIURL, urls.APIURL, fmt.Sprintf("epos-opensource docker deploy %s", env.Name), urls.BackofficeURL)
	},
}

func init() {
	DeployCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Pull Docker images before starting")
	DeployCmd.Flags().StringVar(&configFilePath, "config", "", "Path to YAML configuration file")
}
