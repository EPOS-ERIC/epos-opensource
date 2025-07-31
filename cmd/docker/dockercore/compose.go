package dockercore

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/epos-eu/epos-opensource/command"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/joho/godotenv"
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
	display.Step("Pulling images for environment: %s", name)
	if _, err := command.RunCommand(composeCommand(dir, "", "pull"), false); err != nil {
		return fmt.Errorf("pull images failed: %w", err)
	}
	display.Done("Images pulled for environment: %s", name)
	return nil
}

type Urls struct {
	guiURL        string
	apiURL        string
	backofficeURL string
}

// deployStack deploys the stack in the specified directory.
// the deployed stack will use the given ports for the deployment of the services.
// it is the responsibility of the caller to ensure the ports are not used
func deployStack(dir, name string, ports *DeploymentPorts) (*Urls, error) {
	err := os.Setenv("DATAPORTAL_PORT", strconv.Itoa(ports.GUI))
	if err != nil {
		return nil, fmt.Errorf("failed to set env var 'DATAPORTAL_PORT': %w", err)
	}
	err = os.Setenv("GATEWAY_PORT", strconv.Itoa(ports.API))
	if err != nil {
		return nil, fmt.Errorf("failed to set env var 'GATEWAY_PORT': %w", err)
	}
	err = os.Setenv("BACKOFFICE_PORT", strconv.Itoa(ports.Backoffice))
	if err != nil {
		return nil, fmt.Errorf("failed to set env var 'BACKOFFICE_PORT': %w", err)
	}

	display.Step("Deploying stack")
	if _, err := command.RunCommand(composeCommand(dir, name, "up", "-d"), false); err != nil {
		return nil, fmt.Errorf("deployment of stack failed: %w", err)
	}

	display.Done("Deployed environment: %s", name)

	urls, err := buildEnvURLs(dir, ports)
	if err != nil {
		return nil, fmt.Errorf("error building urls for environment '%s': %w", dir, err)
	}
	return &urls, nil
}

// downStack stops the stack running in the given directory
func downStack(dir string, removeVolumes bool) error {
	if removeVolumes {
		_, err := command.RunCommand(composeCommand(dir, "", "down", "-v"), false)
		return err
	}
	_, err := command.RunCommand(composeCommand(dir, "", "down"), false)
	return err
}

type DeploymentPorts struct {
	GUI        int
	API        int
	Backoffice int
}

// ensureFree makes sure that the ports are free and usable. If they are not, new ones are generated and substituted in place
func (d *DeploymentPorts) ensureFree() error {
	ensurePortFree := func(port int, name string) (int, error) {
		isFree, err := common.IsPortFree(port)
		if err != nil {
			display.Warn("error checking availability of port for %s: %d. Continuing anyway", name, port)
		}
		if !isFree {
			freePort, err := common.FindFreePort()
			if err != nil {
				return 0, fmt.Errorf("error getting free port for %s", name)
			}
			display.Info("Port %d for %s is already used, using new random free port: %d", port, name, freePort)
			return freePort, nil
		}
		return port, nil
	}

	var err error

	d.GUI, err = ensurePortFree(d.GUI, "DATAPORTAL_PORT")
	if err != nil {
		return err
	}
	d.API, err = ensurePortFree(d.API, "GATEWAY_PORT")
	if err != nil {
		return err
	}
	d.Backoffice, err = ensurePortFree(d.Backoffice, "BACKOFFICE_PORT")
	if err != nil {
		return err
	}
	return nil
}

func loadPortsFromEnvFile(filePath string) (*DeploymentPorts, error) {
	env, err := godotenv.Read(filepath.Join(filePath, ".env"))
	if err != nil {
		return nil, fmt.Errorf("failed to read .env file at %s: %w", filepath.Join(filePath, ".env"), err)
	}

	ports := &DeploymentPorts{}

	guiPort, ok := env["DATAPORTAL_PORT"]
	if !ok {
		return nil, fmt.Errorf("environment variable DATAPORTAL_PORT is not set")
	}
	ports.GUI, err = strconv.Atoi(guiPort)
	if err != nil {
		return nil, fmt.Errorf("invalid value for port DATAPORTAL_PORT: %s", guiPort)
	}

	apiPort, ok := env["GATEWAY_PORT"]
	if !ok {
		return nil, fmt.Errorf("environment variable GATEWAY_PORT is not set")
	}
	ports.API, err = strconv.Atoi(guiPort)
	if err != nil {
		return nil, fmt.Errorf("invalid value for port GATEWAY_PORT: %s", apiPort)
	}

	backofficePort, ok := env["BACKOFFICE_PORT"]
	if !ok {
		return nil, fmt.Errorf("environment variable BACKOFFICE_PORT is not set")
	}
	ports.Backoffice, err = strconv.Atoi(backofficePort)
	if err != nil {
		return nil, fmt.Errorf("invalid value for port BACKOFFICE_PORT: %s", backofficePort)
	}

	return ports, nil
}
