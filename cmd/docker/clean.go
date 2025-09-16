// Package docker contains the internal functions used by the docker cmd to manage environments
package docker

import (
	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var CleanCmd = &cobra.Command{
	Use:   "clean [env-name]",
	Short: "Removes the volume of the metadata database.",
	Long:  "Stops the metadata database, removes the volume associated to it and re-deloys the container.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := dockercore.Clean(dockercore.CleanOpts{
			Name: name,
		})
		if err != nil {
			display.Error("%v", err)
			return
		}
	},
}
