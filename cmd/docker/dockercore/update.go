// Package dockercore contains the internal functions used by the docker cmd to manage the environments
package dockercore

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/epos-eu/epos-opensource/display"
)

type UpdateOpts struct {
	// Optional. path to an env file to use. If not set the default embedded one will be used
	EnvFile string
	// Optional. path to a docker-compose.yaml file. If not set the default embedded one will be used
	ComposeFile string
	// Required. name of the environment
	Name string
	// Optional. whether to pull the images before deploying or not
	PullImages bool
	// Optional. whether to do a docker compose down before updating the environment. Useful to reset the environment's database
	Force bool
	// Optional. custom ip to use instead of localhost if set
	CustomIP string
}

// Update logic:
// find the old env, if it does not exist give an error
// if it exists, create a copy of it in a tmp dir
// if force is set do a docker compose down on the original env
// then remove the contents of the env dir and create the updated env file and docker-compose
// if pullImages is set before deploying pull the images for the updated env
// deploy the updated compose
// if everything goes right, delete the tmp dir and finish
// else restore the tmp dir, deploy the old restored env and give an error in output
func Update(opts UpdateOpts) (*sqlc.Docker, error) {
	display.Step("Updating environment: %s", opts.Name)

	docker, err := db.GetDockerByName(opts.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting docker environment from db called '%s': %w", opts.Name, err)
	}
	ports := &DeploymentPorts{
		GUI:        int(docker.GuiPort),
		API:        int(docker.ApiPort),
		Backoffice: int(docker.BackofficePort),
	}

	// If it exists, create a copy of it in a tmp dir
	tmpDir, err := common.CreateTmpCopy(docker.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup copy: %w", err)
	}

	handleFailure := func(msg string, mainErr error) (*sqlc.Docker, error) {
		display.Step("Restoring environment from backup")
		if err := common.RemoveEnvDir(docker.Directory); err != nil {
			display.Error("Failed to remove corrupted directory: %v", err)
		}
		if err := common.RestoreTmpDir(tmpDir, docker.Directory); err != nil {
			display.Error("Failed to restore from backup: %v", err)
		} else {
			// we can ignore the resulting urls since they be the same as the original ones
			if _, err := deployStack(docker.Directory, opts.Name, ports, opts.CustomIP); err != nil {
				display.Error("Failed to deploy restored environment: %v", err)
			}
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return nil, fmt.Errorf(msg, mainErr)
	}

	display.Step("Updating stack")

	// If force is set do a docker compose down on the original env
	if opts.Force {
		if err := downStack(docker.Directory, true); err != nil {
			return handleFailure("docker compose down failed: %w", err)
		}
		display.Done("Stopped environment: %s", opts.Name)
	}

	display.Step("Removing old environment directory")

	// Remove the contents of the env dir and create the updated env file and docker-compose
	if err := common.RemoveEnvDir(docker.Directory); err != nil {
		return handleFailure("failed to remove directory %s: %w", fmt.Errorf("%s: %w", docker.Directory, err))
	}

	display.Step("Creating new environment directory")

	// create the directory in the parent
	path := filepath.Dir(docker.Directory)
	dir, err := NewEnvDir(opts.EnvFile, opts.ComposeFile, path, opts.Name)
	if err != nil {
		return handleFailure("failed to prepare environment directory: %w", err)
	}

	display.Done("Updated environment created in directory: %s", dir)

	// If pullImages is set before deploying pull the images for the updated env
	if opts.PullImages {
		if err := pullEnvImages(dir, opts.Name); err != nil {
			display.Error("Pulling images failed: %v", err)
			return handleFailure("pulling images failed: %w", err)
		}
	}

	// Deploy the updated compose
	// we can ignore the urls returned because they will be the same as before
	if _, err = deployStack(dir, opts.Name, ports, opts.CustomIP); err != nil {
		display.Error("Deploy failed: %v", err)
		return handleFailure("deploy failed: %w", err)
	}

	// only repopulate the ontologies if the database has been cleaned
	if opts.Force {
		if err := common.PopulateOntologies(docker.ApiUrl); err != nil {
			display.Error("error initializing the ontologies in the environment: %v", err)
			return handleFailure("error initializing the ontologies in the environment: %w", err)
		}
	}

	// If everything goes right, delete the tmp dir and finish
	if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
		display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
	}

	return docker, nil
}
