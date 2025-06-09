package internal

import (
	_ "embed"
	"epos-cli/common"
	"fmt"
)

func Update(envFile, composeFile, path, name string, force, pullImages bool) error {
	common.PrintStep("Updating environment: %s", name)
	dir, err := GetEnvDir(path, name)
	if err != nil {
		return fmt.Errorf("failed to get environment directory: %w", err)
	}

	common.PrintDone("Environment found in dir: %s", dir)
	common.PrintStep("Updating stack")

	if force {
		if err := downStack(dir); err != nil {
			return fmt.Errorf("docker compose down failed: %w", err)
		}

		common.PrintDone("Stopped environment: %s", name)
	}
	common.PrintStep("Removing old env dir...")

	if err := removeEnvDir(dir); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dir, err)
	}
	common.PrintStep("Creating new env dir...")

	dir, err = NewEnvDir(envFile, composeFile, path, name)
	if err != nil {
		return fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	common.PrintDone("Updated environment created in dir: %s", dir)

	if pullImages {
		if err := pullEnvImages(dir, name); err != nil {
			common.PrintError("Pulling images failed: %v", err)
			if err := removeEnvDir(dir); err != nil {
				return fmt.Errorf("error deleting environment %s: %w", dir, err)
			}
			return err
		}
	}

	if err := deployStack(dir, name); err != nil {
		common.PrintError("%v", err)
		if err := removeEnvDir(dir); err != nil {
			return fmt.Errorf("error deleting updated environment %s: %w", dir, err)
		}
		return err
	}

	return nil
}
