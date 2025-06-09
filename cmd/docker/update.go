package docker

import (
	"epos-cli/cmd/docker/internal"
	"epos-cli/common"

	"github.com/spf13/cobra"
)

var force bool

var UpdateCmd = &cobra.Command{
	Use:   "update <name of an existing environment>",
	Short: "docker update cmd TODO",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := internal.Update(envFile, composeFile, path, name, force, pullImages)
		if err != nil {
			common.PrintError("%v", err)
			return
		}
	},
}

func init() {
	UpdateCmd.Flags().StringVarP(&path, "path", "p", "", "Custom path for the creation of the dir with the env and compose files")
	UpdateCmd.Flags().BoolVarP(&force, "force", "f", false, "Force the updating of an environment by putting it down before deploying it again")
	UpdateCmd.Flags().BoolVarP(&pullImages, "update-images", "u", false, "If set the images for the environment will be pulled before deploying")
}
