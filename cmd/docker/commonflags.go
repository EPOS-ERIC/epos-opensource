package docker

import "github.com/spf13/cobra"

var (
	envFile     string
	path        string
	composeFile string
)

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Environment variable file, use default if not provided")
	cmd.Flags().StringVarP(&path, "path", "p", "", "Custom path for the creation of the dir with the env and compose files")
	cmd.Flags().StringVarP(&composeFile, "compose-file", "c", "", "Docker compose file, use default if not provided")
}
