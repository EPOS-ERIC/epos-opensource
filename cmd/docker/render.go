package docker

import (
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
	"github.com/spf13/cobra"
)

var renderOutputPath string

var RenderCmd = &cobra.Command{
	Use:   "render [env-name]",
	Short: "Render Docker environment configuration files",
	Long: `Render .env and docker-compose.yaml files from YAML configuration.

The command loads a YAML configuration file, renders the templates, and creates
the environment directory with .env and docker-compose.yaml files.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) > 0 && args[0] != "" {
			name = args[0]
		}

		var cfg *config.EnvConfig
		var err error
		if configFilePath != "" {
			cfg, err = config.LoadConfig(configFilePath)
			if err != nil {
				display.Error("Failed to load config: %v", err)
				os.Exit(1)
			}
		}

		outputPaths, err := docker.Render(docker.RenderOpts{
			Name:       name,
			Config:     cfg,
			OutputPath: renderOutputPath,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.Done("Environment rendered successfully at:")

		for _, path := range outputPaths {
			display.Done("\t%s", path)
		}

		display.Info("You can now deploy this environment with 'epos-opensource docker deploy [env-name] --config <path-to-config.yaml>'")
	},
}

func init() {
	RenderCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "Path to YAML configuration file")
	RenderCmd.Flags().StringVarP(&renderOutputPath, "output", "o", "", "Output directory for environment files")
}
