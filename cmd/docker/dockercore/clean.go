package dockercore

import (
	_ "embed"
	"fmt"
	"os/exec"

	"github.com/epos-eu/epos-opensource/command"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/validate"
)

type CleanOpts struct {
	// Required. name of the environment
	Name string
}

func Clean(opts CleanOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid clean parameters: %w", err)
	}

	docker, err := db.GetDockerByName(opts.Name)
	if err != nil {
		return fmt.Errorf("failed to get docker info for %s: %w", opts.Name, err)
	}
	ports := &DeploymentPorts{
		GUI:        int(docker.GuiPort),
		API:        int(docker.ApiPort),
		Backoffice: int(docker.BackofficePort),
	}
	metadataContainer := fmt.Sprintf("%s-metadata-database", opts.Name)
	backOfficeContainer := fmt.Sprintf("%s-backoffice-service", opts.Name)
	ingestorContainer := fmt.Sprintf("%s-ingestor-service", opts.Name)
	externalAccContainer := fmt.Sprintf("%s-external-access-service", opts.Name)
	resorceContainer := fmt.Sprintf("%s-resources-service", opts.Name)
	volumeName := fmt.Sprintf("%s_psqldata", opts.Name)

	if _, err := command.RunCommand(exec.Command("docker", "stop", metadataContainer), false); err != nil {
		return fmt.Errorf("failed to stop %s: %w", metadataContainer, err)
	}

	if _, err := command.RunCommand(exec.Command("docker", "rm", "-v", metadataContainer), false); err != nil {
		return fmt.Errorf("failed to remove %s: %w", metadataContainer, err)
	}

	if _, err := command.RunCommand(exec.Command("docker", "volume", "rm", volumeName), false); err != nil {
		return fmt.Errorf("failed to remove volume %s: %w", volumeName, err)
	}

	if _, err := command.RunCommand(exec.Command("docker", "stop", backOfficeContainer), false); err != nil {
		return fmt.Errorf("failed to stop %s: %w", backOfficeContainer, err)
	}

	if _, err := command.RunCommand(exec.Command("docker", "stop", ingestorContainer), false); err != nil {
		return fmt.Errorf("failed to stop %s: %w", ingestorContainer, err)
	}

	if _, err := command.RunCommand(exec.Command("docker", "stop", externalAccContainer), false); err != nil {
		return fmt.Errorf("failed to stop %s: %w", externalAccContainer, err)
	}

	if _, err := command.RunCommand(exec.Command("docker", "stop", resorceContainer), false); err != nil {
		return fmt.Errorf("failed to stop %s: %w", externalAccContainer, err)
	}

	if _, err := deployStack(docker.Directory, opts.Name, ports, ""); err != nil {
		return fmt.Errorf("failed to redeploy the metadata database for %s: %w", opts.Name, err)
	}

	if err := common.PopulateOntologies(docker.ApiUrl); err != nil {
		return fmt.Errorf("failed to populate base ontologies in environment %s: %w", opts.Name, err)
	}

	return nil
}

func (c *CleanOpts) Validate() error {
	if err := validate.EnvironmentExistsDocker(c.Name); err != nil {
		return fmt.Errorf("no environment with the name '%s' exists: %w", c.Name, err)
	}

	return nil
}
