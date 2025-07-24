package k8score

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"
)

func Delete(name string) error {
	display.Step("Deleting environment: %s", name)

	env, err := db.GetKubernetesByName(name)
	if err != nil {
		return fmt.Errorf("error getting kubernetes environment from db called '%s': %w", name, err)
	}

	display.Step("Deleting namespace")

	if err := deleteNamespace(name, env.Context); err != nil {
		return fmt.Errorf("error deleting namespace %s, %w", name, err)
	}

	display.Done("Deleted namespace %s", name)

	if err := common.RemoveEnvDir(env.Directory); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
	}

	err = db.DeleteKubernetes(name)
	if err != nil {
		return fmt.Errorf("failed to delete kubernetes %s (dir: %s) in db: %w", name, env.Directory, err)
	}

	display.Done("Deleted environment: %s", name)

	return nil
}
