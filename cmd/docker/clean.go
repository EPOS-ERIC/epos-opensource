package docker

import (
	"fmt"
	"os"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/spf13/cobra"
)

var CleanCmd = &cobra.Command{
	Use:   "clean [env-name]",
	Short: "Clean the data of an environment.",
	Long:  "Clean the data of an environment without redeploying. After clean all custom data populated in the environment will be lost. This action is not reversible.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		docker, err := dockercore.Clean(dockercore.CleanOpts{
			Name: name,
		})
		if err != nil {
			display.Error("%v", err)
			os.Exit(1)
		}

		display.Urls(docker.GuiUrl, docker.ApiUrl, docker.BackofficeUrl, fmt.Sprintf("epos-opensource docker clean %s", name))
	},
}
