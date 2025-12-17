package docker

import (
	"strings"

	"github.com/epos-eu/epos-opensource/db"
	"github.com/spf13/cobra"
)

func validArgsFunction(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if strings.Contains(cmd.Use, "populate") {
		if len(args) == 0 {
			docker, err := db.GetAllDocker()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			var matches []string
			for _, d := range docker {
				if strings.HasPrefix(d.Name, toComplete) {
					matches = append(matches, d.Name)
				}
			}

			return matches, cobra.ShellCompDirectiveNoFileComp
		} else {
			return nil, cobra.ShellCompDirectiveDefault
		}
	} else if strings.Contains(cmd.Use, "delete") {
		if len(args) == 0 {
			docker, err := db.GetAllDocker()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			var matches []string
			for _, d := range docker {
				if strings.HasPrefix(d.Name, toComplete) {
					matches = append(matches, d.Name)
				}
			}

			return matches, cobra.ShellCompDirectiveNoFileComp
		} else {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	} else if strings.Contains(cmd.Use, "clean") {
		docker, err := db.GetAllDocker()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var matches []string
		for _, d := range docker {
			if strings.HasPrefix(d.Name, toComplete) {
				matches = append(matches, d.Name)
			}
		}

		return matches, cobra.ShellCompDirectiveNoFileComp
	} else {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
