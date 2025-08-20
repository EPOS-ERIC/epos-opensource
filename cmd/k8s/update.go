package k8s

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/cmd/k8s/k8score"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var force bool

var UpdateCmd = &cobra.Command{
	Use:   "update [env-name] [flags]",
	Short: "Update and redeploy an existing Kubernetes environment",
	Long:  "Recreates the specified environment with updated configuration or manifests. Optionally deletes and recreates the namespace if --force is used. Ensures rollback if the update fails.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		k, err := k8score.Update(envFile, manifestsDir, name, host, force)
		if err != nil {
			display.Error("%v", err)
			return
		}

		display.Urls(k.GuiUrl, k.ApiUrl, k.BackofficeUrl, fmt.Sprintf("epos-opensource docker update %s", name))
	},
}

func init() {
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Remove the current containers before redeploying")
	UpdateCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Path to the environment variables file (.env)")
	UpdateCmd.Flags().StringVarP(&manifestsDir, "manifests-dir", "m", "", "Path to the directory containing the manifests files")
	UpdateCmd.Flags().StringVar(&host, "host", "", "host (either IP or hostname) to use for exposing the environment. If not set the nginx ingress controller IP is used by default")
}
