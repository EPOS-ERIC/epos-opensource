package k8s

import (
	"os"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:               "delete [env-name...]",
	Short:             "Removes K8s environments and all their namespaces.",
	Long:              "Deletes the K8s environments by removing the namespaces and all of their associated resources. This action is irreversible.",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: validArgsFunction,
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0:]

		if !deleteForce {
			envList := strings.Join(name, ", ")
			display.Warn("This will permanently delete the following environment(s): %s", envList)
			display.Warn("All namespaces and associated resources will be removed. This action cannot be undone.")
			confirmed, err := common.Confirm("Are you sure you want to continue? (y/n):")
			if err != nil {
				display.Error("Failed to read confirmation: %v", err)
				os.Exit(1)
			}
			if !confirmed {
				display.Info("Delete operation cancelled.")
				return
			}
		}

		err := k8s.Delete(k8s.DeleteOpts{
			Name:    name,
			Context: context,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	DeleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Force delete without confirmation prompt")
	DeleteCmd.Flags().StringVar(&context, "context", "", "Kubectl context to use. Uses current context if not set")
}
