package k8s

import (
	"os"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	// TODO: clarify help on what it really does
	Use:   "list",
	Short: "List installed K8s environments.",
	Run: func(cmd *cobra.Command, args []string) {
		contexts := []string{}
		if context == "" {
			c, err := common.GetKubeContexts()
			if err != nil {
				display.Error("Failed to list kubectl contexts: %v", err)
				os.Exit(1)
			}

			contexts = append(contexts, c...)
		} else {
			contexts = append(contexts, context)
		}

		envs := []k8s.Env{}
		for _, context := range contexts {
			kubeEnvs, err := k8s.List(context)
			if err != nil {
				display.Warn("Failed to get environments in context '%s': %v", context, err)
			}

			envs = append(envs, kubeEnvs...)
		}

		rows := make([][]any, len(envs))
		for i, env := range envs {
			urls, err := env.BuildEnvURLs()
			if err != nil {
				display.Error("Failed to build environment URLs for %s: %v", env.Name, err)
				os.Exit(1)
			}

			var backofficeURL string
			if urls.BackofficeURL != nil {
				backofficeURL = *urls.BackofficeURL
			}

			rows[i] = []any{env.Name, env.Context, urls.GUIURL, urls.APIURL, backofficeURL}
		}

		headers := []string{"Name", "Context", "GUI URL", "API URL", "Backoffice URL"}
		display.InfraList(rows, headers, "Installed K8s environments")
	},
}

func init() {
	ListCmd.Flags().StringVar(&context, "context", "", "kubectl context used for the environment deployment. Uses current if not set")
}
