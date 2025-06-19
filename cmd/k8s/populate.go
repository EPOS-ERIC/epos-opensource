package k8s

import (
	"epos-opensource/cmd/k8s/internal"
	"epos-opensource/common"
	"fmt"

	"github.com/spf13/cobra"
)

var PopulateCmd = &cobra.Command{
	Use:   "populate [name] [ttlDir]",
	Short: "populates an existing kubernetes environment",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		ttlDir := args[1]

		portalURL, gatewayURL, err := internal.Populate(path, name, ttlDir)
		if err != nil {
			common.PrintError("%v", err)
			return
		}

		common.PrintUrls(portalURL, gatewayURL, fmt.Sprintf("epos-opensource  kubernetes deploy %s", name))
	},
}

func init() {
	PopulateCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files if not default")
}
