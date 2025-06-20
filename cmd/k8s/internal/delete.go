package internal

import (
	"github.com/epos-eu/epos-opensource/common"
	"fmt"
)

func Delete(customPath, name string) error {
	common.PrintStep("Deleting environment: %s", name)

	dir, err := common.GetEnvDir(customPath, name, pathPrefix)
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

	common.PrintDone("Deleted environment: %s", name)

	return nil
}
