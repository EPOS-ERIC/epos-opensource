package completion

import (
	"strings"

	"github.com/spf13/cobra"
)

// ValueGetter is a function type for retrieving completion values.
type ValueGetter func() ([]string, error)

// SharedValuesCompletion provides completion for a single set of values.
func SharedValuesCompletion(toComplete string, valueGetter ValueGetter) ([]string, cobra.ShellCompDirective) {
	values, err := valueGetter()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return filterCompletionValues(values, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// SharedValidArgs provides completion for environment names based on the command use.
func SharedValidArgs(cmd *cobra.Command, args []string, toComplete string, valueGetter ValueGetter) ([]string, cobra.ShellCompDirective) {
	switch {
	case strings.Contains(cmd.Use, "populate"):
		if len(args) == 0 {
			return SharedValuesCompletion(toComplete, valueGetter)
		}
		return nil, cobra.ShellCompDirectiveDefault
	case strings.Contains(cmd.Use, "delete"):
		if len(args) == 0 {
			return SharedValuesCompletion(toComplete, valueGetter)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	case strings.Contains(cmd.Use, "clean"):
		return SharedValuesCompletion(toComplete, valueGetter)
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func filterCompletionValues(values []string, toComplete string) []string {
	var matches []string
	for _, value := range values {
		if strings.HasPrefix(value, toComplete) {
			matches = append(matches, value)
		}
	}

	return matches
}
