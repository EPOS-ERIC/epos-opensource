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
	Use:   "deploy <env-name>",
	Short: "Deploy a new environment.",
	Long:  "Deploy a new environment. Creates a new namespace and deploys the EPOS services to it. Uses the default configuration unless --config is set.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := loadConfigIfProvided(configFilePath)
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		if cfg == nil {
			cfg = config.GetDefaultConfig()
		}

		cfg.Name = name

		env, err := k8s.Deploy(k8s.DeployOpts{
			Context: context,
			Timeout: timeout,
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
	addContextFlag(DeployCmd)
	DeployCmd.Flags().StringVar(&configFilePath, "config", "", "Path to YAML configuration file")
	DeployCmd.Flags().DurationVar(&timeout, "timeout", 0, "Operation timeout (default: 5m)")
}
