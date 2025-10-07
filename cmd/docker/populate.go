package docker

import (
	"fmt"
	"os"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/spf13/cobra"
)

var PopulateCmd = &cobra.Command{
	Use:   "populate [env-name] [ttl-paths...]",
	Short: "Ingest TTL files or example data into an environment.",
	Long: `Populate an existing environment with all *.ttl files found in the specified directories (recursively),
or ingest the files directly if individual file paths are provided.
Multiple directories and/or files can be provided and will be processed in order.`,
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

		d, err := dockercore.Populate(dockercore.PopulateOpts{
			TTLDirs:          ttlPaths,
			Name:             name,
			Parallel:         parallel,
			PopulateExamples: populateExamples,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.Urls(d.GuiUrl, d.ApiUrl, d.BackofficeUrl, fmt.Sprintf("epos-opensource docker populate %s", name))
	},
}

func init() {
	PopulateCmd.Flags().IntVarP(&parallel, "parallel", "p", 1, "Number of parallel uploads to perform when ingesting TTL files")
	PopulateCmd.Flags().BoolVar(&populateExamples, "example", false, "Populate the environment with example data")
}
