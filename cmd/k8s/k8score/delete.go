package k8score

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
)

func Delete(name string) error {
	common.PrintStep("Deleting environment: %s", name)

	dir, err := common.GetEnvDir(name, platform)
	if err != nil {
		return fmt.Errorf("failed to resolve environment directory: %w", err)
	}

	kubeEnv, err := db.GetKubernetesByName(name)
	if err != nil || kubeEnv == nil {
		return fmt.Errorf("failed to get kubernetes context for environment %s: %w", name, err)
	}
	context := kubeEnv.Context

	common.PrintInfo("Environment directory: %s", dir)
	common.PrintStep("Deleting namespace")

	if err := deleteNamespace(name, context); err != nil {
		return fmt.Errorf("error deleting namespace %s, %w", name, err)
	}

	common.PrintDone("Deleted namespace %s", name)

	if err := common.RemoveEnvDir(dir); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dir, err)
	}

	err = db.DeleteKubernetes(name)
	if err != nil {
		return fmt.Errorf("failed to delete kubernetes %s (dir: %s) in db: %w", name, dir, err)
	}

	common.PrintDone("Deleted environment: %s", name)

	return nil
}
