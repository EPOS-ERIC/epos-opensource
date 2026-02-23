package docker

import (
	"fmt"
	"os/exec"

	"github.com/EPOS-ERIC/epos-opensource/command"
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/db"
	"github.com/EPOS-ERIC/epos-opensource/display"
)

// CleanOpts defines inputs for Clean.
type CleanOpts struct {
	// Required. name of the environment
	Name string
}

// Clean removes runtime data from an existing Docker environment and restarts required services.
func Clean(opts CleanOpts) (*Env, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid clean parameters: %w", err)
	}

	display.Step("Cleaning data in environment: %s", opts.Name)

	env, err := GetEnv(opts.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to load docker environment %s: %w", opts.Name, err)
	}

	urls, err := env.BuildEnvURLs()
	if err != nil {
		return nil, fmt.Errorf("failed to build environment URLs: %w", err)
	}

	// TODO: check if the logic here can be simplified now that we use restart in the depends_on in the docker compose

	metadataContainer := fmt.Sprintf("%s-metadata-database", opts.Name)
	backOfficeContainer := fmt.Sprintf("%s-backoffice-service", opts.Name)
	ingestorContainer := fmt.Sprintf("%s-ingestor-service", opts.Name)
	externalAccContainer := fmt.Sprintf("%s-external-access-service", opts.Name)
	resourcesContainer := fmt.Sprintf("%s-resources-service", opts.Name)
	volumeName := fmt.Sprintf("%s_psqldata", opts.Name)

	handleFailure := func(msg string, mainErr error) (*Env, error) {
		display.Error("Unexpected error: %v", mainErr)
		display.Warn("Clean failed, trying to restore services")
		display.Step("Attempting recovery")

		if err := deployStack(false, &env.EnvConfig); err != nil {
			return nil, fmt.Errorf("unable to recover from error: %w", err)
		}

		display.Done("Recovery succeeded")

		return nil, fmt.Errorf(msg, mainErr)
	}

	display.Step("Cleaning database volume")
	display.Debug("stopping metadata container: %s", metadataContainer)

	if _, err := command.RunCommand(exec.Command("docker", "stop", metadataContainer), false); err != nil {
		return handleFailure("failed to stop metadata container: %w", fmt.Errorf("%s: %w", metadataContainer, err))
	}

	display.Debug("removing metadata container: %s", metadataContainer)

	if _, err := command.RunCommand(exec.Command("docker", "rm", "-v", metadataContainer), false); err != nil {
		return handleFailure("failed to remove metadata container: %w", fmt.Errorf("%s: %w", metadataContainer, err))
	}

	display.Debug("removing database volume: %s", volumeName)

	if _, err := command.RunCommand(exec.Command("docker", "volume", "rm", volumeName), false); err != nil {
		return handleFailure("failed to remove volume: %w", fmt.Errorf("%s: %w", volumeName, err))
	}

	display.Done("Database volume cleaned")

	if err := db.DeleteIngestedFilesByEnvironment(opts.Name); err != nil {
		return handleFailure("failed to clear ingested files tracking: %w", err)
	}

	display.Debug("cleared ingested file records for environment: %s", opts.Name)
	display.Step("Restarting services")
	display.Debug("stopping container: %s", backOfficeContainer)

	if _, err := command.RunCommand(exec.Command("docker", "stop", backOfficeContainer), false); err != nil {
		return handleFailure("failed to stop container: %w", fmt.Errorf("%s: %w", backOfficeContainer, err))
	}

	display.Debug("stopping container: %s", ingestorContainer)

	if _, err := command.RunCommand(exec.Command("docker", "stop", ingestorContainer), false); err != nil {
		return handleFailure("failed to stop container: %w", fmt.Errorf("%s: %w", ingestorContainer, err))
	}

	display.Debug("stopping container: %s", externalAccContainer)

	if _, err := command.RunCommand(exec.Command("docker", "stop", externalAccContainer), false); err != nil {
		return handleFailure("failed to stop container: %w", fmt.Errorf("%s: %w", externalAccContainer, err))
	}

	display.Debug("stopping container: %s", resourcesContainer)

	if _, err := command.RunCommand(exec.Command("docker", "stop", resourcesContainer), false); err != nil {
		return handleFailure("failed to stop container: %w", fmt.Errorf("%s: %w", resourcesContainer, err))
	}

	display.Debug("redeploying stack for environment: %s", env.Name)

	if err := deployStack(false, &env.EnvConfig); err != nil {
		return handleFailure("failed to restart the environment: %w", err)
	}

	display.Done("Services restarted successfully")

	if err := common.PopulateOntologies(urls.APIURL); err != nil {
		return handleFailure("failed to populate base ontologies in environment: %w", err)
	}

	display.Debug("repopulated base ontologies from: %s", urls.APIURL)
	display.Done("Cleaned environment: %s", opts.Name)

	return env, nil
}

// Validate checks CleanOpts and ensures the target environment exists.
func (c *CleanOpts) Validate() error {
	display.Debug("name: %s", c.Name)

	if err := EnsureEnvironmentExists(c.Name); err != nil {
		return fmt.Errorf("no environment with the name '%s' exists: %w", c.Name, err)
	}

	return nil
}
