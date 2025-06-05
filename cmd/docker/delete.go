package docker

import (
	"epos-cli/cmd/docker/internal"
	"epos-cli/common"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "docker delete cmd TODO",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := internal.Delete(path, name)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&path, "path", "p", "", "Custom path for the creation of the dir with the env and compose files")
}
