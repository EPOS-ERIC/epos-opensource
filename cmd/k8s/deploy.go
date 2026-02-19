// Package k8s contains the cobra cmd implementation for the k8s management
package k8s

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy [env-name]",
	Short: "Create and deploy a new K8s environment in a dedicated namespace.",
	Long: `Sets up a new K8s environment in a fresh namespace, applying all required manifests and configuration. Fails if the namespace already exists.
NOTE: to execute the deploy it will try to use port-forwarding to the cluster. If that fails it will retry using the external api.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		display.Debug("configFilePath: %s", configFilePath)
		display.Debug("context: %s", context)

		var cfg *config.Config
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

		env, err := k8s.Deploy(k8s.DeployOpts{
			Context: context,
			Config:  cfg,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		URLs, err := env.BuildEnvURLs()
		if err != nil {
			display.Error("Failed to build environment URLs: %v", err)
			os.Exit(1)
		}

		display.URLs(URLs.GUIURL, URLs.APIURL, fmt.Sprintf("epos-opensource k8s deploy %s", name), URLs.BackofficeURL)
	},
}

func init() {
	DeployCmd.Flags().StringVar(&context, "context", "", "kubectl context used for the environment deployment. Uses current if not set")
	DeployCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "Path to YAML configuration file")
}
