package k8s

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/release"
)

func ReleaseToConfig(rel *release.Release) (*config.EnvConfig, error) {
	if rel == nil {
		return nil, fmt.Errorf("release is nil")
	}

	yamlBytes, err := yaml.Marshal(rel.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config to yaml: %w", err)
	}

	var envConfig config.EnvConfig
	if err := yaml.Unmarshal(yamlBytes, &envConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &envConfig, nil
}

func ReleaseToEnv(rel *release.Release, context string) (*Env, error) {
	if rel == nil {
		return nil, fmt.Errorf("release is nil")
	}

	config, err := ReleaseToConfig(rel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert release to config: %w", err)
	}

	return &Env{
		EnvConfig: *config,
		Name:      config.Name,
		Context:   context,
	}, nil
}
