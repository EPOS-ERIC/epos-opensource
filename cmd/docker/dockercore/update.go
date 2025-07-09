// Package dockercore contains the internal functions used by the docker cmd to manage the environments
package dockercore

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
)

// Update logic:
// find the old env, if it does not exist give an error
// if it exists, create a copy of it in a tmp dir
// if force is set do a docker compose down on the original env
// then remove the contents of the env dir and create the updated env file and docker-compose
// if pullImages is set before deploying pull the images for the updated env
// deploy the updated compose
// if everything goes right, delete the tmp dir and finish
// else restore the tmp dir, deploy the old restored env and give an error in output
func Update(envFile, composeFile, name string, force, pullImages bool) (*db.Docker, error) {
	common.PrintStep("Updating environment: %s", name)

	docker, err := db.GetDockerByName(name)
	if err != nil {
		return nil, fmt.Errorf("error getting docker environment from db called '%s': %w", name, err)
	}

	// If it exists, create a copy of it in a tmp dir
	tmpDir, err := common.CreateTmpCopy(docker.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup copy: %w", err)
	}

	handleFailure := func(msg string, mainErr error) (*db.Docker, error) {
		common.PrintStep("Restoring environment from backup")
		if err := common.RemoveEnvDir(docker.Directory); err != nil {
			common.PrintError("Failed to remove corrupted directory: %v", err)
		}
		if err := common.RestoreTmpDir(tmpDir, docker.Directory); err != nil {
			common.PrintError("Failed to restore from backup: %v", err)
		} else {
			if err := deployStack(docker.Directory, name); err != nil {
				common.PrintError("Failed to deploy restored environment: %v", err)
			}
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return nil, fmt.Errorf(msg, mainErr)
	}

	common.PrintStep("Updating stack")

	// If force is set do a docker compose down on the original env
	if force {
		if err := downStack(docker.Directory, true); err != nil {
			return handleFailure("docker compose down failed: %w", err)
		}
		common.PrintDone("Stopped environment: %s", name)
	}

	common.PrintStep("Removing old environment directory")

	// Remove the contents of the env dir and create the updated env file and docker-compose
	if err := common.RemoveEnvDir(docker.Directory); err != nil {
		return handleFailure("failed to remove directory %s: %w", fmt.Errorf("%s: %w", docker.Directory, err))
	}

	common.PrintStep("Creating new environment directory")

	// create the directory in the parent
	path := filepath.Dir(docker.Directory)
	dir, err := NewEnvDir(envFile, composeFile, path, name)
	if err != nil {
		return handleFailure("failed to prepare environment directory: %w", err)
	}

	common.PrintDone("Updated environment created in directory: %s", dir)

	// If pullImages is set before deploying pull the images for the updated env
	if pullImages {
		if err := pullEnvImages(dir, name); err != nil {
			common.PrintError("Pulling images failed: %v", err)
			return handleFailure("pulling images failed: %w", err)
		}
	}

	// Deploy the updated compose
	if err := deployStack(dir, name); err != nil {
		common.PrintError("Deploy failed: %v", err)
		return handleFailure("deploy failed: %w", err)
	}

	// only repopulate the ontologies if the database has been cleaned
	if force {
		if err := common.PopulateOntologies(docker.ApiUrl); err != nil {
			common.PrintError("error initializing the ontologies in the environment: %v", err)
			return handleFailure("error initializing the ontologies in the environment: %w", err)
		}
	}

	// If everything goes right, delete the tmp dir and finish
	if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
		common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
	}

	return docker, nil
}
