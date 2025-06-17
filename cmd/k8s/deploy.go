package k8s

import (
	"epos-cli/cmd/k8s/internal"
	"epos-cli/common"
	"fmt"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy <name to give to the environment>",
	Short: "Create a new environment using Kubernetes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		portalURL, gatewayURL, err := internal.Deploy(envFile, manifestsDir, path, name)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
		common.PrintUrls(portalURL, gatewayURL, fmt.Sprintf("epos-cli kubernetes deploy %s", name))
	},
}

func init() {
	DeployCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Path to the environment variables file (.env)")
	DeployCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files")
	DeployCmd.Flags().StringVarP(&manifestsDir, "manifests-dir", "m", "", "Path to the directory containing the manifests files")
}
