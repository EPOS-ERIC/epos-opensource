package k8s

import (
	"os"

	"github.com/epos-eu/epos-opensource/cmd/k8s/k8score"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [env-name...]",
	Short: "Removes Kubernetes environments and all their namespaces.",
	Long:  "Deletes the Kubernetes environments by removing the namespaces and all of their associated resources. This action is irreversible.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0:]

		err := k8score.Delete(k8score.DeleteOpts{
			Name: name,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}
	},
}
