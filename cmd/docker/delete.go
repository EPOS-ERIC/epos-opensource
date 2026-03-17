// Package docker contains the internal functions used by the docker cmd to manage environments
package docker

import (
	"os"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:               "delete <env-name>...",
	Short:             "Delete one or more environments.",
	Long:              "Delete one or more environments. Removes the Docker Compose environment, including its containers, volumes, and tracked metadata. Prompts for confirmation unless --force is set.",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: validArgsFunction,
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0:]

		if !deleteForce {
			envList := strings.Join(name, ", ")
			display.Warn("This will permanently delete the following environment(s): %s", envList)
			display.Warn("All containers, volumes, and associated resources will be removed. This action cannot be undone.")
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

		err := docker.Delete(docker.DeleteOpts{
			Name: name,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	DeleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip the confirmation prompt")
}
