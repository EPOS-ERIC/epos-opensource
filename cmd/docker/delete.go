package docker

import (
	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/common"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [env-name]",
	Short: "Stop and remove a Docker Compose environment",
	Long:  "Deletes the Docker Compose environment with the given name.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := dockercore.Delete(name)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
	},
}
