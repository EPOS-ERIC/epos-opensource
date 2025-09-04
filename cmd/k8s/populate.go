package k8s

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/k8s/k8score"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/spf13/cobra"
)

var PopulateCmd = &cobra.Command{
	Use:   "populate [env-name] [ttl-paths...]",
	Short: "Ingest TTL files from directories or files into an environment.",
	Long: `Populate an existing environment with all *.ttl files found in the specified directories (recursively),
or ingest the files directly if individual file paths are provided.
Multiple directories and/or files can be provided and will be processed in order.
NOTE: To execute the population it will try to use port-forwarding to the cluster. If that fails it will retry using the external API.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		ttlPaths := args[1:]

		k, err := k8score.Populate(k8score.PopulateOpts{
			TTLDirs:  ttlPaths,
			Name:     name,
			Parallel: parallel,
		})
		if err != nil {
			display.Error("%v", err)
			return
		}
		display.Urls(k.GuiUrl, k.ApiUrl, k.BackofficeUrl, fmt.Sprintf("epos-opensource kubernetes populate %s", name))
	},
}

func init() {
	PopulateCmd.Flags().IntVarP(&parallel, "parallel", "p", 1, "Number of parallel uploads to perform when ingesting TTL files")
}
