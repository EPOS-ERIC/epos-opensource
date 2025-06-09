package internal

import (
	_ "embed"
	"epos-cli/common"
	"fmt"
	"os/exec"
)

func Deploy(envFile, composeFile, path, name string, pullImages bool) error {
	common.PrintStep("Creating environment: %s", name)
	dir, err := NewEnvDir(envFile, composeFile, path, name)
	if err != nil {
		return fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	common.PrintDone("Environment created in dir: %s", dir)

	if pullImages {
		common.PrintStep("Pulling images for environment %s", name)
		cmd := exec.Command("docker", "compose", "pull")
		cmd.Dir = dir
		if err := common.RunCommand(cmd); err != nil {
			common.PrintError("Pulling images failed: %v", err)
			common.PrintStep("Deleting environment dir: %s", dir)

			if err := DeleteEnvDir(dir); err != nil {
				return fmt.Errorf("error deleting environment %s: %w", dir, err)
			}
			common.PrintDone("Deleted invalid environment at: %s", dir)
			return err
		}
		common.PrintDone("Images pulled for environment %s", name)
	}

	common.PrintStep("Deploying stack")

	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = dir
	cmd.Env = append(cmd.Env, "ENV_NAME="+name)
	if err := common.RunCommand(cmd); err != nil {
		common.PrintError("Deploying stack failed: %v", err)
		common.PrintStep("Deleting environment dir: %s", dir)

		cmd := exec.Command("docker", "compose", "down")
		cmd.Dir = dir
		err = common.RunCommand(cmd)
		if err != nil {
			common.PrintWarn("Docker compose down failed, there may be some dangling env: %v", err)
		}

		if err := DeleteEnvDir(dir); err != nil {
			return fmt.Errorf("error deleting environment %s: %w", dir, err)
		}
		common.PrintDone("Deleted invalid environment at: %s", dir)
		common.PrintError("Deploying stack failed")
		return err
	}

	common.PrintDone("Deployed environment: %s", name)

	return nil
}
