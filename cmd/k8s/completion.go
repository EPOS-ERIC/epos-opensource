package k8s

import (
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
	"github.com/spf13/cobra"
)

func validArgsFunction(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return common.SharedValidArgsFunction(cmd, args, toComplete, func() ([]string, error) {
		envs, err := k8s.List(context)
		if err != nil {
			return nil, err
		}

		names := make([]string, len(envs))
		for i, k := range envs {
			names[i] = k.Name
		}

		return names, nil
	})
}
