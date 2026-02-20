package k8s

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// List returns all EPOS-managed K8s environments visible in the given context.
func List(context string) ([]Env, error) {
	settings := cli.New()
	settings.KubeContext = context

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		"", // all namespaces
		"secret",
		func(format string, v ...any) { display.Info(format, v...) },
	); err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %w", err)
	}

	client := action.NewList(actionConfig)

	releases, err := client.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get helm chart: %w", err)
	}

	envConfigs := []Env{}

	for _, release := range releases {
		if release.Chart.Metadata.Annotations["managed-by"] != "epos" {
			continue
		}

		envConfig, err := ReleaseToEnv(release, context)
		if err != nil {
			return nil, fmt.Errorf("failed to convert release to config: %w", err)
		}

		envConfigs = append(envConfigs, *envConfig)
	}

	return envConfigs, nil
}
