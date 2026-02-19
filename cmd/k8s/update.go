package k8s

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"

	"github.com/spf13/cobra"
)

var (
	force bool
	reset bool
)

var UpdateCmd = &cobra.Command{
	Use:               "update [env-name]",
	Short:             "Update and redeploy an existing K8s environment.",
	Long:              "Recreates the specified environment with updated configuration or manifests. Optionally deletes and recreates the namespace if --force is used. Ensures rollback if the update fails.",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: validArgsFunction,
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		display.Debug("name: %v", name)
		display.Debug("args: %v", args)
		display.Debug("reset: %v", reset)
		display.Debug("force: %v", force)

		// TODO: this is reused in many cli commands, abstract it?
		var cfg *config.Config
		var err error
		if configFilePath != "" {
			cfg, err = config.LoadConfig(configFilePath)
			if err != nil {
				display.Error("Failed to load config: %v", err)
				os.Exit(1)
			}
		}

		env, err := k8s.Update(k8s.UpdateOpts{
			Force:      force,
			Reset:      reset,
			OldEnvName: name,
			NewConfig:  cfg,
			Context:    context,
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
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Delete and recreate the namespace and all resources before redeploying.")
	UpdateCmd.Flags().BoolVarP(&reset, "reset", "r", false, "Reset .env and manifests to embedded versions")
	UpdateCmd.Flags().StringVar(&configFilePath, "config", "", "Path to YAML configuration file")
	UpdateCmd.Flags().StringVar(&context, "context", "", "kubectl context used for the environment deployment. Uses current if not set")
}
