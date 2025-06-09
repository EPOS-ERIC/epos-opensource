package internal

import (
	"epos-cli/common"
	"fmt"
	"os/exec"
)

// composeCommand creates a docker compose command configured with the given directory and environment name
func composeCommand(dir, name string, args ...string) *exec.Cmd {
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Dir = dir
	if name != "" {
		cmd.Env = append(cmd.Env, "ENV_NAME="+name)
	}
	return cmd
}

// pullEnvImages pulls docker images for the environment with custom messages
func pullEnvImages(dir, name string) error {
	common.PrintStep("Pulling images for environment %s", name)
	if err := common.RunCommand(composeCommand(dir, "", "pull")); err != nil {
		return fmt.Errorf("pull images failed: %w", err)
	}
	common.PrintDone("Images pulled for environment %s", name)
	return nil
}

// deployStack deploys the stack in the specified directory
func deployStack(dir, name string) error {
	common.PrintStep("Deploying stack")
	if err := common.RunCommand(composeCommand(dir, name, "up", "-d")); err != nil {
		return fmt.Errorf("deploy stack failed: %w", err)
	}
	common.PrintDone("Deployed environment: %s", name)
	return nil
}

// downStack stops the stack running in the given directory
func downStack(dir string) error {
	return common.RunCommand(composeCommand(dir, "", "down"))
}

// removeEnvDir deletes the environment directory with logs
func removeEnvDir(dir string) error {
	common.PrintStep("Deleting environment dir: %s", dir)
	if err := DeleteEnvDir(dir); err != nil {
		return err
	}
	common.PrintDone("Deleted environment at: %s", dir)
	return nil
}
