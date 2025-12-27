package dockercore

import (
	_ "embed"
	"fmt"
	"os/exec"

	"github.com/EPOS-ERIC/epos-opensource/command"
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/db"
	"github.com/EPOS-ERIC/epos-opensource/db/sqlc"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/validate"
)

type CleanOpts struct {
	// Required. name of the environment
	Name string
}

func Clean(opts CleanOpts) (*sqlc.Docker, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid clean parameters: %w", err)
	}

	display.Step("Cleaning data in environment: %s", opts.Name)

	docker, err := db.GetDockerByName(opts.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker info for %s: %w", opts.Name, err)
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
	resourcesContainer := fmt.Sprintf("%s-resources-service", opts.Name)
	volumeName := fmt.Sprintf("%s_psqldata", opts.Name)

	handleFailure := func(msg string, mainErr error) (*sqlc.Docker, error) {
		display.Error("Unexpected error: %v", mainErr)

		display.Step("Attempting recovery")
		if _, err := deployStack(docker.Directory, opts.Name, ports, ""); err != nil {
			return nil, fmt.Errorf("unable to recover from error: %w", err)
		}

		display.Done("Recovery succeeded")

		return nil, fmt.Errorf(msg, mainErr)
	}

	display.Step("Cleaning database volume")

	if _, err := command.RunCommand(exec.Command("docker", "stop", metadataContainer), false); err != nil {
		return handleFailure("failed to stop metadata container: %w", fmt.Errorf("%s: %w", metadataContainer, err))
	}

	if _, err := command.RunCommand(exec.Command("docker", "rm", "-v", metadataContainer), false); err != nil {
		return handleFailure("failed to remove metadata container: %w", fmt.Errorf("%s: %w", metadataContainer, err))
	}

	if _, err := command.RunCommand(exec.Command("docker", "volume", "rm", volumeName), false); err != nil {
		return handleFailure("failed to remove volume: %w", fmt.Errorf("%s: %w", volumeName, err))
	}

	display.Done("Database volume cleaned")

	// Clear ingested files tracking
	if err := db.DeleteIngestedFilesByEnvironment("docker", opts.Name); err != nil {
		return handleFailure("failed to clear ingested files tracking: %w", err)
	}

	display.Step("Restarting services")

	if _, err := command.RunCommand(exec.Command("docker", "stop", backOfficeContainer), false); err != nil {
		return handleFailure("failed to stop container: %w", fmt.Errorf("%s: %w", backOfficeContainer, err))
	}

	if _, err := command.RunCommand(exec.Command("docker", "stop", ingestorContainer), false); err != nil {
		return handleFailure("failed to stop container: %w", fmt.Errorf("%s: %w", ingestorContainer, err))
	}

	if _, err := command.RunCommand(exec.Command("docker", "stop", externalAccContainer), false); err != nil {
		return handleFailure("failed to stop container: %w", fmt.Errorf("%s: %w", externalAccContainer, err))
	}

	if _, err := command.RunCommand(exec.Command("docker", "stop", resourcesContainer), false); err != nil {
		return handleFailure("failed to stop container: %w", fmt.Errorf("%s: %w", resourcesContainer, err))
	}

	if _, err := deployStack(docker.Directory, opts.Name, ports, ""); err != nil {
		return handleFailure("failed to restart the environment: %w", err)
	}

	display.Done("Services restarted successfully")

	if err := common.PopulateOntologies(docker.ApiUrl); err != nil {
		return handleFailure("failed to populate base ontologies in environment: %w", err)
	}

	display.Done("Cleaned environment: %s", opts.Name)

	return docker, nil
}

func (c *CleanOpts) Validate() error {
	if err := validate.EnvironmentExistsDocker(c.Name); err != nil {
		return fmt.Errorf("no environment with the name '%s' exists: %w", c.Name, err)
	}

	return nil
}
