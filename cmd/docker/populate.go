package docker

import (
	"fmt"
	"github.com/epos-eu/epos-opensource/cmd/docker/internal"
	"github.com/epos-eu/epos-opensource/common"

	"github.com/spf13/cobra"
)

var ttlDirPath string

var PopulateCmd = &cobra.Command{
	Use:   "populate [env-name] [ttl-directory]",
	Short: "Ingest TTL files into an environment",
	Long:  "Populate an existing environment with all *.ttl files found in the specified directory (recursively).",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		ttlDir := args[1]

		portalURL, gatewayURL, err := internal.Populate(path, name, ttlDir)
		if err != nil {
			common.PrintError("%v", err)
			return
		}

		common.PrintUrls(portalURL, gatewayURL, fmt.Sprintf("epos-opensource docker populate %s", name))
	},
}

func init() {
	PopulateCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files if not default")
}
