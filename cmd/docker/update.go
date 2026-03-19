package docker

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"

	"github.com/spf13/cobra"
)

var (
	force bool
	reset bool
)

var UpdateCmd = &cobra.Command{
	Use:   "update <env-name>",
	Short: "Update an existing environment.",
	Long:  "Update an existing environment. Updates the deployed environment using the current applied configuration or a file passed with --config. Use --reset to start from the default configuration, --force to recreate containers, or --update-images to pull images before starting.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		var cfg *config.EnvConfig
		var err error
		if configFilePath != "" {
			cfg, err = config.LoadConfig(configFilePath)
			if err != nil {
				display.Error("Failed to load config: %v", err)
				os.Exit(1)
			}
		}

		env, err := docker.Update(docker.UpdateOpts{
			PullImages: pullImages,
			Force:      force,
			Reset:      reset,
			OldEnvName: name,
			NewConfig:  cfg,
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

		display.URLs(urls.GUIURL, urls.APIURL, fmt.Sprintf("epos-opensource docker update %s", name), urls.BackofficeURL)
	},
}

func init() {
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Recreate the environment by removing current containers first")
	UpdateCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Pull Docker images before starting")
	UpdateCmd.Flags().BoolVar(&reset, "reset", false, "Use the embedded default config")
	UpdateCmd.Flags().StringVar(&configFilePath, "config", "", "Path to YAML configuration file")
}
