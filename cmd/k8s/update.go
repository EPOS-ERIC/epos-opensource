package k8s

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"

	"github.com/spf13/cobra"
)

var (
	force bool
	reset bool
)

var UpdateCmd = &cobra.Command{
	Use:               "update <env-name>",
	Short:             "Update an existing environment.",
	Long:              "Update an existing environment. Updates the deployed environment using the current applied configuration or a file passed with --config. Use --reset to start from the default configuration or --force to delete and recreate the environment.",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: validArgsFunction,
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := loadConfigIfProvided(configFilePath)
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		env, err := k8s.Update(k8s.UpdateOpts{
			Force:      force,
			Reset:      reset,
			OldEnvName: name,
			NewConfig:  cfg,
			Context:    context,
			Timeout:    timeout,
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

		display.URLs(URLs.GUIURL, URLs.APIURL, fmt.Sprintf("epos-opensource k8s update %s", name), URLs.BackofficeURL)
	},
}

func init() {
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Reinstall from scratch by deleting the namespace first")
	UpdateCmd.Flags().BoolVarP(&reset, "reset", "r", false, "Use the embedded default config")
	UpdateCmd.Flags().StringVar(&configFilePath, "config", "", "Path to YAML configuration file")
	UpdateCmd.Flags().StringVar(&context, "context", "", "kubectl context to use (default: current context)")
	UpdateCmd.Flags().DurationVar(&timeout, "timeout", 0, "Operation timeout (default: 5m)")
}
