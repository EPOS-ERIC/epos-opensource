package internal

import (
	_ "embed"
	"epos-cli/common"
	"os/exec"
)

func Deploy(envFile, composeFile, path, name string) error {
	dir, err := NewEnvDir(envFile, composeFile, path, name)
	if err != nil {
		return err
	}

	common.PrintInfo("Using environment in dir: %s", dir)
	common.PrintStep("Deploying stack")

	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = dir
	err = common.RunCommand(cmd)
	if err != nil {
		return err
	}

	common.PrintDone("Deployed environment: %s", name)

	return nil
}
