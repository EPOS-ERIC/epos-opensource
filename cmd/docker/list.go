package docker

import (
	"net/url"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/db"
	"github.com/EPOS-ERIC/epos-opensource/display"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Docker environments.",
	Run: func(cmd *cobra.Command, args []string) {
		dockerEnvs, err := db.GetAllDocker()
		if err != nil {
			display.Error("Failed to list Docker environments: %v", err)
			os.Exit(1)
		}

		rows := make([][]any, len(dockerEnvs))
		for i, dockerEnv := range dockerEnvs {
			apiURL, err := url.JoinPath(dockerEnv.ApiUrl, "ui")
			if err != nil {
				display.Warn("Could not construct gateway URL: %v", err)
				apiURL = dockerEnv.ApiUrl
			}

			var backofficeURL string
			if dockerEnv.BackofficeUrl != nil {
				u, err := url.JoinPath(*dockerEnv.BackofficeUrl, "home")
				if err != nil {
					display.Warn("Could not construct backoffice URL: %v", err)
					backofficeURL = *dockerEnv.BackofficeUrl
				} else {
					backofficeURL = u
				}
			}

			rows[i] = []any{dockerEnv.Name, dockerEnv.Directory, dockerEnv.GuiUrl, apiURL, backofficeURL}
		}

		headers := []string{"Name", "Directory", "GUI URL", "API URL", "Backoffice URL"}
		display.InfraList(rows, headers, "Installed Docker environments")
	},
}
