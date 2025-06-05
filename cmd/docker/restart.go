package docker

import (
	"epos-cli/cmd/docker/internal"
	"epos-cli/common"

	"github.com/spf13/cobra"
)

var RestartCmd = &cobra.Command{
	Use:   "restart <name>",
	Short: "docker restart cmd TODO",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := internal.Restart(path, name)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
	},
}

func init() {
	RestartCmd.Flags().StringVarP(&path, "path", "p", "", "Custom path for the creation of the dir with the env and compose files")
}
