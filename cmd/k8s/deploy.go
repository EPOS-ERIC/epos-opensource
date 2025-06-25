// Package k8s contains the cobra cmd implementation for the k8s management
package k8s

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/k8s/internal"
	"github.com/epos-eu/epos-opensource/common"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy [env-name]",
	Short: "Create and deploy a new Kubernetes environment in a dedicated namespace",
	Long:  "Sets up a new Kubernetes environment in a fresh namespace, applying all required manifests and configuration. Fails if the namespace already exists.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		portalURL, gatewayURL, err := internal.Deploy(envFile, manifestsDir, path, name)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
		common.PrintUrls(portalURL, gatewayURL, fmt.Sprintf("epos-opensource  kubernetes deploy %s", name))
	},
}

func init() {
	DeployCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Path to the environment variables file (.env)")
	DeployCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files")
	DeployCmd.Flags().StringVarP(&manifestsDir, "manifests-dir", "m", "", "Path to the directory containing the manifests files")
}
