// Package docker contains the internal functions used by the docker cmd to manage environments
package docker

import (
	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [env-name]",
	Short: "Stop and remove a Docker Compose environment.",
	Long:  "Deletes the Docker Compose environment with the given name.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := dockercore.Delete(dockercore.DeleteOpts{
			Name: name,
		})
		if err != nil {
			display.Error("%v", err)
			return
		}
	},
}
