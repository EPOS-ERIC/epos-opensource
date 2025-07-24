package docker

import (
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "list installed docker environments",
	Run: func(cmd *cobra.Command, args []string) {
		dockerEnvs, err := db.GetAllDocker()
		if err != nil {
			return
		}

		display.DockerList(dockerEnvs, "installed docker environments")
	},
}
