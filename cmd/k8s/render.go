package k8s

import (
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
	"github.com/spf13/cobra"
)

var renderOutputPath string

var RenderCmd = &cobra.Command{
	Use:   "render [env-name]",
	Short: "Render k8s environment configuration files",
	Long: `TODO: Render the helm chart files from YAML configuration.

The command loads a YAML configuration file, renders the templates, and creates
the environment directory with .env and k8s-compose.yaml files.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		var cfg *config.Config
		var err error
		if configFilePath != "" {
			cfg, err = config.LoadConfig(configFilePath)
			if err != nil {
				display.Error("Failed to load config: %v", err)
				os.Exit(1)
			}
		}
		outputPaths, err := k8s.Render(k8s.RenderOpts{
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
		display.Info("You can now use 'k8s deploy --env %s --k8s-compose %s' to deploy the environment", outputPaths[0], outputPaths[1])
	},
}

func init() {
	RenderCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "Path to YAML configuration file")
	RenderCmd.Flags().StringVarP(&renderOutputPath, "output", "o", "", "Output directory for environment files")
}
