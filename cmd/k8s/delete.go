package k8s

import (
	"github.com/epos-eu/epos-opensource/cmd/k8s/k8score"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [env-name]",
	Short: "Remove a Kubernetes environment and its namespace",
	Long:  "Deletes the specified Kubernetes environment by removing its namespace and all associated resources. This action is irreversible.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := k8score.Delete(k8score.DeleteOpts{
			Name: name,
		})
		if err != nil {
			display.Error("%v", err)
			return
		}
	},
}
