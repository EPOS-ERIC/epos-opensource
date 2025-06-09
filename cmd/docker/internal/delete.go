package internal

import (
	_ "embed"
	"epos-cli/common"
	"fmt"
)

func Delete(customPath, name string) error {
	dir, err := GetEnvDir(customPath, name)
	if err != nil {
		return fmt.Errorf("failed to resolve environment directory: %w", err)
	}

	common.PrintInfo("Using environment in dir: %s", dir)
	common.PrintStep("Stopping stack...")

	if err := downStack(dir); err != nil {
		return fmt.Errorf("docker compose down failed: %w", err)
	}

	common.PrintDone("Stopped environment: %s", name)

	if err := removeEnvDir(dir); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dir, err)
	}

	return nil
}
