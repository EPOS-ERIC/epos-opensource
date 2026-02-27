package k8s

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
	"github.com/spf13/cobra"
)

// loadConfigIfProvided loads a YAML config file when a path is provided.
func loadConfigIfProvided(configPath string) (*config.Config, error) {
	if configPath == "" {
		return nil, nil
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %q: %w", configPath, err)
	}

	return cfg, nil
}

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
