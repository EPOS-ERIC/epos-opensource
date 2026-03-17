package docker

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	"github.com/spf13/cobra"
)

var PopulateCmd = &cobra.Command{
	Use:               "populate <env-name> [ttl-paths...]",
	Short:             "Load TTL data into an environment.",
	Long:              "Load TTL data into an environment. Imports .ttl files from the given files or directories, or loads bundled example data with --example. Pass at least one TTL path unless --example is set.",
	ValidArgsFunction: validArgsFunction,
	Args: func(cmd *cobra.Command, args []string) error {
		if populateExamples {
			if len(args) < 1 {
				display.Error("requires environment name when using --example")
				return fmt.Errorf("requires environment name when using --example")
			}
			return nil
		}
		if len(args) < 2 {
			display.Error("requires environment name and at least one TTL path (or use --example flag)")
			return fmt.Errorf("requires environment name and at least one TTL path (or use --example flag)")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		ttlPaths := args[1:]

		env, err := docker.Populate(docker.PopulateOpts{
			TTLDirs:          ttlPaths,
			Name:             name,
			Parallel:         parallel,
			PopulateExamples: populateExamples,
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

		display.URLs(urls.GUIURL, urls.APIURL, fmt.Sprintf("epos-opensource docker populate %s", name), urls.BackofficeURL)
	},
}

func init() {
	PopulateCmd.Flags().IntVarP(&parallel, "parallel", "p", 1, "Parallel TTL uploads (1-20)")
	PopulateCmd.Flags().BoolVar(&populateExamples, "example", false, "Load bundled example data")
}
