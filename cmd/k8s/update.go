package k8s

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/cmd/k8s/k8score"
	"github.com/EPOS-ERIC/epos-opensource/display"

	"github.com/spf13/cobra"
)

var (
	force bool
	reset bool
)

var UpdateCmd = &cobra.Command{
	Use:               "update [env-name]",
	Short:             "Update and redeploy an existing K8s environment.",
	Long:              "Recreates the specified environment with updated configuration or manifests. Optionally deletes and recreates the namespace if --force is used. Ensures rollback if the update fails.",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: validArgsFunction,
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		k, err := k8score.Update(k8score.UpdateOpts{
			EnvFile:     envFile,
			ManifestDir: manifestsDir,
			Name:        name,
			Force:       force,
			CustomHost:  host,
			Reset:       reset,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.Urls(k.GuiUrl, k.ApiUrl, k.BackofficeUrl, fmt.Sprintf("epos-opensource k8s update %s", name))
	},
}

func init() {
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Delete and recreate the namespace and all resources before redeploying.")
	UpdateCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Path to the environment variables file (.env)")
	UpdateCmd.Flags().StringVarP(&manifestsDir, "manifests-dir", "m", "", "Path to the directory containing the manifests files")
	UpdateCmd.Flags().StringVar(&host, "host", "", "Host (either IP or hostname) to use for exposing the environment. If not set the nginx ingress controller IP is used by default")
	UpdateCmd.Flags().BoolVarP(&reset, "reset", "r", false, "Reset .env and manifests to embedded versions")
}
