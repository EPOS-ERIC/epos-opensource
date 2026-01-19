package docker

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore"
	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore/config"
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

		d, err := dockercore.Update(dockercore.UpdateOpts{
			EnvFile:     envFilePath,
			ComposeFile: composeFilePath,
			Name:        name,
			PullImages:  pullImages,
			Force:       force,
			Reset:       reset,
			Config:      cfg,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.URLs(d.GuiUrl, d.ApiUrl, fmt.Sprintf("epos-opensource docker update %s", name), d.BackofficeUrl)
	},
}

func init() {
	UpdateCmd.Flags().StringVar(&envFilePath, "env-file", "", "Path to the environment variables file (.env)")
	UpdateCmd.Flags().StringVar(&composeFilePath, "compose-file", "", "Path to the Docker Compose file")
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Remove the current containers before redeploying")
	UpdateCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "Download Docker images before starting")
	UpdateCmd.Flags().BoolVar(&reset, "reset", false, "Reset the environment config to the embedded defaults")
	UpdateCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "Path to YAML configuration file")
}
