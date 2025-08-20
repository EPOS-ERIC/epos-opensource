// Package k8score contains the internal functions used by the kubernetes cmd to manage the environments
package k8score

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/epos-eu/epos-opensource/display"
)

// Update logic:
// find the old env, if it does not exist give an error
// if it exists, create a copy of it in a tmp dir
// if force is set delete the kubernetes namespace for the original env
// then remove the contents of the env dir and create the updated env file and kubernetes manifests
// if pullImages is set before deploying pull the images for the updated env
// deploy the updated manifests
// if everything goes right, delete the tmp dir and finish
// else restore the tmp dir, deploy the old restored env and give an error in output
func Update(envFile, composeFile, name, customIP string, force bool) (*sqlc.Kubernetes, error) {
	display.Step("Updating environment: %s", name)

	kube, err := db.GetKubernetesByName(name)
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes environment from db called '%s': %w", name, err)
	}

	display.Info("Environment directory: %s", kube.Directory)

	// If it exists, create a copy of it in a tmp dir
	tmpDir, err := common.CreateTmpCopy(kube.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup copy: %w", err)
	}

	// Cleanup function to restore from tmp if needed
	restoreFromTmp := func() error {
		display.Step("Restoring environment from backup")
		if err := common.RemoveEnvDir(kube.Directory); err != nil {
			display.Error("Failed to remove corrupted directory: %v", err)
		}
		if err := common.RestoreTmpDir(tmpDir, kube.Directory); err != nil {
			return fmt.Errorf("failed to restore from backup: %w", err)
		}
		if err := deployManifests(kube.Directory, name, true, kube.Context, kube.Protocol); err != nil {
			return fmt.Errorf("failed to deploy restored environment: %w", err)
		}
		return nil
	}

	display.Step("Updating stack")

	// If force is set delete the original env
	if force {
		if err := deleteNamespace(name, kube.Context); err != nil {
			if restoreErr := restoreFromTmp(); restoreErr != nil {
				display.Error("Restore failed: %v", restoreErr)
			}
			if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
				display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
			}
			return nil, fmt.Errorf("kubernetes namespace deletion failed: %w", err)
		}
		display.Done("Stopped environment: %s", name)
	}

	display.Step("Removing old environment directory")

	// Remove the contents of the env dir and create the updated env file and manifests
	if err := common.RemoveEnvDir(kube.Directory); err != nil {
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			display.Error("Restore failed: %v", restoreErr)
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return nil, fmt.Errorf("failed to remove directory %s: %w", kube.Directory, err)
	}

	display.Step("Creating new environment directory")

	// create the directory in the parent
	path := filepath.Dir(kube.Directory)
	dir, err := NewEnvDir(envFile, composeFile, path, name, kube.Context, kube.Protocol, customIP)
	if err != nil {
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			display.Error("Restore failed: %v", restoreErr)
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return nil, fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	display.Done("Updated environment created in directory: %s", dir)

	// Deploy the updated manifests
	if err := deployManifests(dir, name, force, kube.Context, kube.Protocol); err != nil {
		display.Error("Deploy failed: %v", err)
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			display.Error("Restore failed: %v", restoreErr)
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return nil, err
	}

	// only repopulate the ontologies if the database has been cleaned
	if force {
		if err := common.PopulateOntologies(kube.ApiUrl); err != nil {
			display.Error("error initializing the ontologies in the environment: %v", err)
			if restoreErr := restoreFromTmp(); restoreErr != nil {
				display.Error("Restore failed: %v", restoreErr)
			}
			if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
				display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
			}
			return nil, fmt.Errorf("error initializing the ontologies in the environment: %w", err)
		}
	}

	// If everything goes right, delete the tmp dir and finish
	if err := common.RemoveTmpDir(tmpDir); err != nil {
		display.Warn("Failed to cleanup tmp dir: %v", err)
		// Don't return error as the main operation succeeded
	}

	return kube, nil
}
