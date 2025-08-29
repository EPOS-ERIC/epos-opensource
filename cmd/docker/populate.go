package docker

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/spf13/cobra"
)

var PopulateCmd = &cobra.Command{
	Use:   "populate [env-name] [ttl-paths...]",
	Short: "Ingest TTL files from one or more directories or individual files into an environment",
	Long: `Populate an existing environment with all *.ttl files found in the specified directories (recursively),
or ingest the files directly if individual file paths are provided.
Multiple directories and/or files can be provided and will be processed in order.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		ttlPaths := args[1:]
		d, err := dockercore.Populate(dockercore.PopulateOpts{
			TTLDirs:  ttlPaths,
			Name:     name,
			Parallel: parallel,
		})
		if err != nil {
			display.Error("%v", err)
			return
		}
		display.Urls(d.GuiUrl, d.ApiUrl, d.BackofficeUrl, fmt.Sprintf("epos-opensource docker populate %s", name))
	},
}

func init() {
	PopulateCmd.Flags().IntVarP(&parallel, "parallel", "p", 1, "Number of parallel uploads to perform when ingesting TTL files. Default is 1")
}
