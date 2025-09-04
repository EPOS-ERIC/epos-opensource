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
	"github.com/epos-eu/epos-opensource/validate"
)

type UpdateOpts struct {
	// Optional. path to an env file to use. If not set the default embedded one will be used
	EnvFile string
	// Optional. path to a directory that will contain the custom manifest to use for the deploy
	ManifestDir string
	// Required. name of the environment
	Name string
	// Optional. whether to do a docker compose down before updating the environment. Useful to reset the environment's database
	Force bool
	// Optional. custom ip to use instead of localhost if set
	CustomHost string
}

// Update logic:
// find the old env, if it does not exist give an error
// if it exists, create a copy of it in a tmp dir
// if force is set delete the kubernetes namespace for the original env
// then remove the contents of the env dir and create the updated env file and kubernetes manifests
// if pullImages is set before deploying pull the images for the updated env
// deploy the updated manifests
// if everything goes right, delete the tmp dir and finish
// else restore the tmp dir, deploy the old restored env and give an error in output
func Update(opts UpdateOpts) (*sqlc.Kubernetes, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters for update command: %w", err)
	}
	display.Step("Updating environment: %s", opts.Name)

	kube, err := db.GetKubernetesByName(opts.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes environment from db called '%s': %w", opts.Name, err)
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
		if err := deployManifests(kube.Directory, opts.Name, true, kube.Context, kube.Protocol); err != nil {
			return fmt.Errorf("failed to deploy restored environment: %w", err)
		}
		return nil
	}

	display.Step("Updating stack")

	// If force is set delete the original env
	if opts.Force {
		if err := deleteNamespace(opts.Name, kube.Context); err != nil {
			if restoreErr := restoreFromTmp(); restoreErr != nil {
				display.Error("Restore failed: %v", restoreErr)
			}
			if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
				display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
			}
			return nil, fmt.Errorf("kubernetes namespace deletion failed: %w", err)
		}
		display.Done("Stopped environment: %s", opts.Name)
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
	dir, err := NewEnvDir(opts.EnvFile, opts.ManifestDir, path, opts.Name, kube.Context, kube.Protocol, opts.CustomHost)
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
	if err := deployManifests(dir, opts.Name, opts.Force, kube.Context, kube.Protocol); err != nil {
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
	if opts.Force {
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

func (u *UpdateOpts) Validate() error {
	if err := validate.EnvironmentExistsK8s(u.Name); err != nil {
		return fmt.Errorf("no environment with name '%s' exists: %w", u.Name, err)
	}

	if err := validate.CustomHost(u.CustomHost); err != nil {
		return fmt.Errorf("custom host '%s' is invalid: %w ", u.CustomHost, err)
	}

	if err := validate.IsFile(u.EnvFile); err != nil {
		return fmt.Errorf("the path to .env '%s' is not a file: %w", u.EnvFile, err)
	}

	if err := validate.PathExists(u.ManifestDir); err != nil {
		return fmt.Errorf("the manifest directory path '%s' is not a valid path: %w", u.ManifestDir, err)
	}

	return nil
}
