package k8s

import (
	"fmt"
	"os"

	"github.com/epos-eu/epos-opensource/cmd/k8s/k8score"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/spf13/cobra"
)

var PopulateCmd = &cobra.Command{
	Use:   "populate [env-name] [ttl-paths...]",
	Short: "Ingest TTL files or example data into an environment.",
	Long: `Populate an existing environment with all *.ttl files found in the specified directories (recursively),
or ingest the files directly if individual file paths are provided.
Multiple directories and/or files can be provided and will be processed in order.
NOTE: To execute the population it will try to use port-forwarding to the cluster. If that fails it will retry using the external API.`,
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

		k, err := k8score.Populate(k8score.PopulateOpts{
			TTLDirs:          ttlPaths,
			Name:             name,
			Parallel:         parallel,
			PopulateExamples: populateExamples,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.Urls(k.GuiUrl, k.ApiUrl, k.BackofficeUrl, fmt.Sprintf("epos-opensource kubernetes populate %s", name))
	},
}

func init() {
	PopulateCmd.Flags().IntVarP(&parallel, "parallel", "p", 1, "Number of parallel uploads to perform when ingesting TTL files")
	PopulateCmd.Flags().BoolVar(&populateExamples, "example", false, "Populate the environment with example data")
}
