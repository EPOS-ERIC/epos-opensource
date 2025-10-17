// Package docker contains the internal functions used by the docker cmd to manage environments
package docker

import (
	"os"
	"strings"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [env-name...]",
	Short: "Stop and remove Docker Compose environments.",
	Long:  "Deletes Docker Compose environments with the given names. This action is irreversible.",
	Args:  cobra.MinimumNArgs(1),
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

		err := dockercore.Delete(dockercore.DeleteOpts{
			Name: name,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	DeleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Force delete without confirmation prompt")
}
