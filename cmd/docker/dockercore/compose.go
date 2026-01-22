package dockercore

import (
	"fmt"
	"os/exec"

	"github.com/EPOS-ERIC/epos-opensource/command"
	"github.com/EPOS-ERIC/epos-opensource/display"
)

// composeCommand creates a docker compose command configured with the given directory and environment name
func composeCommand(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Dir = dir
	return cmd
}

// pullEnvImages pulls docker images for the environment with custom messages
// TODO make this take just the image struct and pull those, not using the env file
func pullEnvImages(dir, name string) error {
	display.Step("Pulling images for environment: %s", name)
	if _, err := command.RunCommand(composeCommand(dir, "pull"), false); err != nil {
		return fmt.Errorf("pull images failed: %w", err)
	}
	display.Done("Images pulled for environment: %s", name)
	return nil
}

// deployStack deploys the stack in the specified directory.
// the deployed stack will use the given ports for the deployment of the services.
// it is the responsibility of the caller to ensure the ports are not used
func deployStack(removeOrphans bool, dir string) error {
	display.Step("Deploying stack")

	args := []string{"up", "-d"}
	if removeOrphans {
		args = append(args, "--remove-orphans")
	}
	if _, err := command.RunCommand(composeCommand(dir, args...), false); err != nil {
		return fmt.Errorf("deployment of stack failed: %w", err)
	}

	display.Done("Stack deployed successfully")

	return nil
}

// downStack stops the stack running in the given directory
func downStack(dir string, removeVolumes bool) error {
	if removeVolumes {
		_, err := command.RunCommand(composeCommand(dir, "down", "-v"), false)
		return err
	}
	_, err := command.RunCommand(composeCommand(dir, "down"), false)
	return err
}

type DeploymentPorts struct {
	GUI        int
	API        int
	Backoffice int
}

// ensureFree makes sure that the ports are free and usable. If they are not, new ones are generated and substituted in place
// func (d *DeploymentPorts) ensureFree() error {
// 	ensurePortFree := func(port int, name string) (int, error) {
// 		isFree, err := common.IsPortFree(port)
// 		if err != nil {
// 			display.Warn("error checking availability of port for %s: %d. Continuing anyway", name, port)
// 		}
// 		if !isFree {
// 			freePort, err := common.FindFreePort()
// 			if err != nil {
// 				return 0, fmt.Errorf("error getting free port for %s", name)
// 			}
// 			display.Info("Port %d for %s is already used, using new random free port: %d", port, name, freePort)
// 			return freePort, nil
// 		}
// 		return port, nil
// 	}
//
// 	var err error
//
// 	d.GUI, err = ensurePortFree(d.GUI, "DATAPORTAL_PORT")
// 	if err != nil {
// 		return err
// 	}
// 	d.API, err = ensurePortFree(d.API, "GATEWAY_PORT")
// 	if err != nil {
// 		return err
// 	}
// 	d.Backoffice, err = ensurePortFree(d.Backoffice, "BACKOFFICE_PORT")
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
