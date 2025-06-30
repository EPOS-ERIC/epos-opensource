package docker

import (
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"

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

		common.PrintDockerList(dockerEnvs, "installed docker environments")
	},
}
