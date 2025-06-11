package docker

import (
	"epos-cli/cmd/docker/internal"
	"epos-cli/common"

	"github.com/spf13/cobra"
)

var ttlDirPath string

var PopulateCmd = &cobra.Command{
	Use:   "populate <name of an existing environment> <path to a directory containing ttl files>",
	Short: "Populate an existing environment with new data from a directory containing *.ttl files. The dir is walked recursively adding all *.ttl files found.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		ttlDir := args[1]

		portalURL, gatewayURL, err := internal.Populate(path, name, ttlDir)
		if err != nil {
			common.PrintError("%v", err)
			return
		}

		common.PrintUrls(portalURL, gatewayURL)
	},
}

func init() {
	PopulateCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files if not default")
}
