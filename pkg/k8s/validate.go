package k8s

import (
	"errors"
	"fmt"
	"slices"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// EnsureEnvironmentDoesNotExist checks that the K8s environment does not exist.
func EnsureEnvironmentDoesNotExist(name, context string) error {
	_, err := GetEnv(name, context)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get environment %q: %w", name, err)
	}

	return fmt.Errorf("an environment with name '%s' already exists", name)
}

// EnsureEnvironmentExists checks that the K8s environment exists.
func EnsureEnvironmentExists(name, context string) error {
	_, err := GetEnv(name, context)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return fmt.Errorf("no environment with name '%s' exists", name)
		}
		return fmt.Errorf("failed to get environment %q: %w", name, err)
	}

	return nil
}

func EnsureContextExists(context string) error {
	contexts, err := common.GetKubeContexts()
	if err != nil {
		return fmt.Errorf("failed to list kubectl contexts: %w", err)
	}

	contextFound := slices.Contains(contexts, context)
	if !contextFound {
		return fmt.Errorf("K8s context %q is not an available context", context)
	}

	return nil
}
