package k8s

import (
	"strings"

	"github.com/epos-eu/epos-opensource/db"
	"github.com/spf13/cobra"
)

func validArgsFunction(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	docker, err := db.GetAllKubernetes()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var matches []string
	for _, k := range docker {
		if strings.HasPrefix(k.Name, toComplete) {
			matches = append(matches, k.Name)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}
