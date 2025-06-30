package k8score

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
)

func Delete(customPath, name string) error {
	common.PrintStep("Deleting environment: %s", name)

	dir, err := common.GetEnvDir(customPath, name, platform)
	if err != nil {
		return fmt.Errorf("failed to resolve environment directory: %w", err)
	}

	common.PrintInfo("Environment directory: %s", dir)
	common.PrintStep("Deleting namespace")

	if err := deleteNamespace(name); err != nil {
		return fmt.Errorf("error deleting namespace %s, %w", name, err)
	}

	common.PrintDone("Deleted namespace %s", name)

	if err := common.RemoveEnvDir(dir); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dir, err)
	}

	err = db.DeleteEnv(name, "kubernetes")
	if err != nil {
		return fmt.Errorf("failed to delete env %s (dir: %s, platform: %s) in db: %w", name, dir, "kubernetes", err)
	}

	common.PrintDone("Deleted environment: %s", name)

	return nil
}
