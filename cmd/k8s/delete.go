package k8s

import (
	"epos-opensource/cmd/k8s/internal"
	"epos-opensource/common"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [env-name]",
	Short: "Stop and remove a Kubernetes environment",
	Long:  "Deletes the Kubernetes environment and namespace with the given name.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := internal.Delete(path, name)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&path, "path", "p", "", "Location for the environment files")
}
