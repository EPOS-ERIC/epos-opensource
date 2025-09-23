// Package docker contains the internal functions used by the docker cmd to manage environments
package docker

import (
	"os"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [env-name...]",
	Short: "Stop and remove Docker Compose environments.",
	Long:  "Deletes Docker Compose environments with the given names.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0:]

		err := dockercore.Delete(dockercore.DeleteOpts{
			Name: name,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}
	},
}
