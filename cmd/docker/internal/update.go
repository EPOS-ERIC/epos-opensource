package internal

import (
	_ "embed"
	"epos-cli/common"
	"fmt"
	"os/exec"
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
		cmd := exec.Command("docker", "compose", "down")
		cmd.Dir = dir
		err = common.RunCommand(cmd)
		if err != nil {
			return fmt.Errorf("docker compose down failed: %w", err)
		}

		common.PrintDone("Stopped environment: %s", name)
	}
	common.PrintStep("Removing old env dir...")

	err = DeleteEnvDir(dir)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dir, err)
	}

	common.PrintDone("Deleted old env dir: %s", dir)
	common.PrintStep("Creating new env dir...")

	dir, err = NewEnvDir(envFile, composeFile, path, name)
	if err != nil {
		return fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	common.PrintDone("Updated environment created in dir: %s", dir)

	if pullImages {
		common.PrintStep("Pulling images for updated environment %s", name)
		cmd := exec.Command("docker", "compose", "pull")
		cmd.Dir = dir
		if err := common.RunCommand(cmd); err != nil {
			common.PrintError("Pulling images failed: %v", err)
			common.PrintStep("Deleting environment dir: %s", dir)

			if err := DeleteEnvDir(dir); err != nil {
				return fmt.Errorf("error deleting environment %s: %w", dir, err)
			}
			common.PrintDone("Deleted invalid updated environment at: %s", dir)
			return err
		}
		common.PrintDone("Images pulled for updated environment %s", name)
	}

	common.PrintStep("Deploying updated stack")

	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = dir
	cmd.Env = append(cmd.Env, "ENV_NAME="+name)
	if err := common.RunCommand(cmd); err != nil {
		common.PrintError("Deploying updated stack failed: %v", err)
		common.PrintStep("Deleting updated environment dir: %s", dir)

		if err := DeleteEnvDir(dir); err != nil {
			return fmt.Errorf("error deleting updated environment %s: %w", dir, err)
		}
		common.PrintDone("Deleted invalid updated environment at: %s", dir)
		return err
	}

	common.PrintDone("Deployed updated environment: %s", name)

	return nil
}
