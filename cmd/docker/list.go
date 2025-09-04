package docker

import (
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Docker environments.",
	Run: func(cmd *cobra.Command, args []string) {
		dockerEnvs, err := db.GetAllDocker()
		if err != nil {
			return
		}

		display.DockerList(dockerEnvs, "Installed Docker environments")
	},
}
