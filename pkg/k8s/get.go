package k8s

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// GetEnv fetches a deployed K8s environment by release name and context.
func GetEnv(name, context string) (*Env, error) {
	settings := cli.New()
	settings.KubeContext = context

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		name,
		"secret",
		func(format string, v ...any) { display.Info(format, v...) },
	); err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %w", err)
	}

	client := action.NewGet(actionConfig)

	rel, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get helm chart: %w", err)
	}

	env, err := EnvFromRelease(rel, context)
	if err != nil {
		return nil, fmt.Errorf("failed to convert release to config: %w", err)
	}

	return env, nil
}
