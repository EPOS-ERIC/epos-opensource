package common

import (
	"strings"

	"github.com/spf13/cobra"
)

// NameGetter is a function type for retrieving environment names.
type NameGetter func() ([]string, error)

// SharedValidArgsFunction provides completion for environment names based on the command use.
func SharedValidArgsFunction(cmd *cobra.Command, args []string, toComplete string, nameGetter NameGetter) ([]string, cobra.ShellCompDirective) {
	if strings.Contains(cmd.Use, "populate") {
		if len(args) == 0 {
			names, err := nameGetter()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			var matches []string
			for _, name := range names {
				if strings.HasPrefix(name, toComplete) {
					matches = append(matches, name)
				}
			}

			return matches, cobra.ShellCompDirectiveNoFileComp
		} else {
			return nil, cobra.ShellCompDirectiveDefault
		}
	} else if strings.Contains(cmd.Use, "delete") {
		if len(args) == 0 {
			names, err := nameGetter()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			var matches []string
			for _, name := range names {
				if strings.HasPrefix(name, toComplete) {
					matches = append(matches, name)
				}
			}

			return matches, cobra.ShellCompDirectiveNoFileComp
		} else {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	} else if strings.Contains(cmd.Use, "clean") {
		names, err := nameGetter()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var matches []string
		for _, name := range names {
			if strings.HasPrefix(name, toComplete) {
				matches = append(matches, name)
			}
		}

		return matches, cobra.ShellCompDirectiveNoFileComp
	} else {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
