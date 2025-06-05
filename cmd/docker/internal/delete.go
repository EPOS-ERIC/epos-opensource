package internal

import (
	_ "embed"
	"epos-cli/common"
	"os/exec"
)

func Delete(customPath, name string) error {
	dir, err := GetEnvDir(customPath, name)
	if err != nil {
		return err
	}

	common.PrintInfo("Using environment in dir: %s", dir)
	common.PrintStep("Stopping stack...")

	cmd := exec.Command("docker", "compose", "down")
	cmd.Dir = dir
	err = common.RunCommand(cmd)
	if err != nil {
		return err
	}

	common.PrintDone("Stopped environment: %s", name)
	common.PrintStep("Removing env dir...")

	err = DeleteEnvDir(dir)
	if err != nil {
		return err
	}

	common.PrintDone("Deleted env dir: %s", dir)

	return nil
}
