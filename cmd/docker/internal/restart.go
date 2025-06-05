package internal

import (
	"epos-cli/common"
	"os/exec"
)

func Restart(customPath, name string) error {
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
	common.PrintStep("Restarting stack...")

	cmd = exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = dir
	err = common.RunCommand(cmd)
	if err != nil {
		return err
	}

	common.PrintDone("Restarted environment: %s", name)

	return nil
}
