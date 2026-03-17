package k8s

import (
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
	"github.com/spf13/cobra"
)

var renderOutputPath string

var RenderCmd = &cobra.Command{
	Use:   "render [env-name]",
	Short: "Render Kubernetes manifests from config.",
	Long:  "Render Kubernetes manifests from config. Generates manifest files locally from the default configuration or a file passed with --config. Writes the rendered files to the current directory or to --output.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) > 0 && args[0] != "" {
			name = args[0]
		}

		cfg, err := loadConfigIfProvided(configFilePath)
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
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
		display.Info("You can now deploy this environment with 'epos-opensource k8s deploy [env-name] --config <path-to-config.yaml>'")
	},
}

func init() {
	RenderCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "Path to YAML configuration file")
	RenderCmd.Flags().StringVarP(&renderOutputPath, "output", "o", "", "Directory for rendered manifests")
}
