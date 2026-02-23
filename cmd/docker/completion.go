package docker

import (
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker"
	"github.com/spf13/cobra"
)

func validArgsFunction(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return common.SharedValidArgsFunction(cmd, args, toComplete, func() ([]string, error) {
		envs, err := docker.List()
		if err != nil {
			return nil, err
		}

		names := make([]string, len(envs))
		for i, d := range envs {
			names[i] = d.Name
		}

		return names, nil
	})
}
