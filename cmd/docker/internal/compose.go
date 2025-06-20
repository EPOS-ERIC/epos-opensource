package internal

import (
	"github.com/epos-eu/epos-opensource/common"
	"fmt"
	"os/exec"
	"strconv"
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
	common.PrintStep("Pulling images for environment: %s", name)
	if err := common.RunCommand(composeCommand(dir, "", "pull"), false); err != nil {
		return fmt.Errorf("pull images failed: %w", err)
	}
	common.PrintDone("Images pulled for environment: %s", name)
	return nil
}

// deployStack deploys the stack in the specified directory
func deployStack(dir, name string) error {
	common.PrintStep("Deploying stack")
	if err := common.RunCommand(composeCommand(dir, name, "up", "-d"), false); err != nil {
		return fmt.Errorf("deploy stack failed: %w", err)
	}
	common.PrintDone("Deployed environment: %s", name)
	return nil
}

// downStack stops the stack running in the given directory
func downStack(dir string, removeVolumes bool) error {
	if removeVolumes {
		return common.RunCommand(composeCommand(dir, "", "down", "-v"), false)
	}
	return common.RunCommand(composeCommand(dir, "", "down"), false)
}

// deployMetadataCache deploys an nginx docker container running a file server exposing a volume
func deployMetadataCache(dir, envName string) (int, error) {
	port, err := common.GetFreePort()
	if err != nil {
		return 0, fmt.Errorf("error getting a free port for the metadata-cache: %w", err)
	}
	cmd := exec.Command(
		"docker",
		"run",
		"-d",
		"--name",
		envName+"-metadata-cache",
		"-p",
		strconv.Itoa(port)+":80",
		"-v",
		dir+":/usr/share/nginx/html",
		"nginx",
	)

	err = common.RunCommand(cmd, false)
	if err != nil {
		return 0, fmt.Errorf("error deploying metadata-cache: %w", err)
	}

	return port, nil
}

// deployMetadataCache removes a deployment of a metadata cache container
func deleteMetadataCache(envName string) error {
	cmd := exec.Command(
		"docker",
		"rm",
		"-f",
		envName+"-metadata-cache",
	)

	if err := common.RunCommand(cmd, false); err != nil {
		return fmt.Errorf("error removing metadata-cache: %w", err)
	}

	return nil
}
