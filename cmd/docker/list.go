package docker

import (
	"net/url"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Docker environments.",
	Run: func(cmd *cobra.Command, args []string) {
		envs, err := docker.List()
		if err != nil {
			display.Error("Failed to list Docker environments: %v", err)
			os.Exit(1)
		}

		rows := make([][]any, len(envs))
		for i, dockerEnv := range envs {
			urls, err := dockerEnv.BuildEnvURLs()
			if err != nil {
				display.Error("Failed to build environment URLs for %s: %v", dockerEnv.Name, err)
				os.Exit(1)
			}

			apiURL, err := url.JoinPath(urls.APIURL, "ui")
			if err != nil {
				display.Warn("Could not construct gateway URL: %v", err)
				apiURL = urls.APIURL
			}

			var backofficeURL string
			if urls.BackofficeURL != nil {
				u, err := url.JoinPath(*urls.BackofficeURL, "home")
				if err != nil {
					display.Warn("Could not construct backoffice URL: %v", err)
					backofficeURL = *urls.BackofficeURL
				} else {
					backofficeURL = u
				}
			}

			rows[i] = []any{dockerEnv.Name, urls.GUIURL, apiURL, backofficeURL}
		}

		headers := []string{"Name", "GUI URL", "API URL", "Backoffice URL"}
		display.InfraList(rows, headers, "Installed Docker environments")
	},
}
