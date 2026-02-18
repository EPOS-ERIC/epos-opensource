// Package k8s contains the internal functions used by the k8s cmd to manage the environments
package k8s

import (
	_ "embed"
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/db/sqlc"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
)

type UpdateOpts struct {
	// Required. name of the environment
	OldEnvName string
	// Optional. do an uninstall then an install (to wipe volumes and stuff) maybe we should rename the flag?
	Force bool
	// Optional. reset the environment config to the embedded defaults maybe we should rename the flag?
	Reset bool
	// New configuration to apply. If nil, preserves existing config
	NewConfig *config.EnvConfig
}

// Update TODO: add docs
func Update(opts UpdateOpts) (*sqlc.K8s, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters for update command: %w", err)
	}

	display.Step("Updating environment: %s", opts.OldEnvName)

	// kube, err := db.GetK8sByName(opts.Name)
	// if err != nil {
	// 	return nil, fmt.Errorf("error getting k8s environment from db called '%s': %w", opts.Name, err)
	// }

	// display.Info("Environment directory: %s", kube.Directory)

	// // If it exists, create a copy of it in a tmp dir
	// tmpDir, err := common.CreateTmpCopy(kube.Directory)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create backup copy: %w", err)
	// }

	// // If no custom files provided and not resetting, use existing files from tmp dir
	// if opts.EnvFile == "" && !opts.Reset {
	// 	opts.EnvFile = filepath.Join(tmpDir, ".env")
	// }
	// if opts.ManifestDir == "" && !opts.Reset {
	// 	opts.ManifestDir = tmpDir
	// }

	// // Cleanup function to restore from tmp if needed
	// restoreFromTmp := func() error {
	// 	display.Step("Restoring environment from backup")
	// 	if err := common.RemoveEnvDir(kube.Directory); err != nil {
	// 		display.Error("Failed to remove corrupted directory: %v", err)
	// 	}
	// 	if err := common.RestoreTmpDir(tmpDir, kube.Directory); err != nil {
	// 		return fmt.Errorf("failed to restore from backup: %w", err)
	// 	}
	// 	// if err := deployManifests(kube.Directory, opts.Name, true, kube.Context, kube.TlsEnabled); err != nil {
	// 	// 	return fmt.Errorf("failed to deploy restored environment: %w", err)
	// 	// }
	// 	return nil
	// }

	display.Step("Updating stack")

	// // If force is set delete the original env
	// if opts.Force {
	// 	if err := deleteNamespace(opts.Name, kube.Context); err != nil {
	// 		if restoreErr := restoreFromTmp(); restoreErr != nil {
	// 			display.Error("Restore failed: %v", restoreErr)
	// 		}
	// 		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
	// 			display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
	// 		}
	// 		return nil, fmt.Errorf("K8s namespace deletion failed: %w", err)
	// 	}
	// 	display.Done("Stopped environment: %s", opts.Name)
	// }

	display.Step("Removing old environment directory")

	// // Remove the contents of the env dir and create the updated env file and manifests
	// if err := common.RemoveEnvDir(kube.Directory); err != nil {
	// 	if restoreErr := restoreFromTmp(); restoreErr != nil {
	// 		display.Error("Restore failed: %v", restoreErr)
	// 	}
	// 	if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
	// 		display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
	// 	}
	// 	return nil, fmt.Errorf("failed to remove directory %s: %w", kube.Directory, err)
	// }

	display.Step("Creating new environment directory")

	// create the directory in the parent
	// path := filepath.Dir(kube.Directory)
	// dir, err := NewEnvDir(opts.EnvFile, opts.ManifestDir, path, opts.Name, kube.Context, kube.Protocol, opts.CustomHost)
	// if err != nil {
	// 	if restoreErr := restoreFromTmp(); restoreErr != nil {
	// 		display.Error("Restore failed: %v", restoreErr)
	// 	}
	// 	if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
	// 		display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
	// 	}
	// 	return nil, fmt.Errorf("failed to prepare environment directory: %w", err)
	// }

	// display.Done("Updated environment created in directory: %s", dir)
	//
	// // Deploy the updated manifests
	// if err := deployManifests(dir, opts.Name, opts.Force, kube.Context, kube.TlsEnabled); err != nil {
	// 	display.Error("Deploy failed: %v", err)
	// 	if restoreErr := restoreFromTmp(); restoreErr != nil {
	// 		display.Error("Restore failed: %v", restoreErr)
	// 	}
	// 	if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
	// 		display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
	// 	}
	// 	return nil, err
	// }

	// // only repopulate the ontologies if the database has been cleaned
	// if opts.Force {
	// 	if err := common.PopulateOntologies(kube.ApiUrl); err != nil {
	// 		display.Error("error initializing the ontologies in the environment: %v", err)
	// 		if restoreErr := restoreFromTmp(); restoreErr != nil {
	// 			display.Error("Restore failed: %v", restoreErr)
	// 		}
	// 		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
	// 			display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
	// 		}
	// 		return nil, fmt.Errorf("error initializing the ontologies in the environment: %w", err)
	// 	}
	// 	if err := db.DeleteIngestedFilesByEnvironment("k8s", opts.Name); err != nil {
	// 		if restoreErr := restoreFromTmp(); restoreErr != nil {
	// 			display.Error("Restore failed: %v", restoreErr)
	// 		}
	// 		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
	// 			display.Error("Failed to cleanup tmp dir: %v", cleanupErr)
	// 		}
	// 		return nil, fmt.Errorf("failed to clear ingested files tracking: %w", err)
	// 	}
	// }

	// // If everything goes right, delete the tmp dir and finish
	// if err := common.RemoveTmpDir(tmpDir); err != nil {
	// 	display.Warn("Failed to cleanup tmp dir: %v", err)
	// 	// Don't return error as the main operation succeeded
	// }

	return nil, nil
}

func (u *UpdateOpts) Validate() error {
	// if err := validate.EnvironmentExistsK8s(u.Name); err != nil {
	// 	return fmt.Errorf("no environment with name '%s' exists: %w", u.Name, err)
	// }
	//
	// if err := validate.CustomHost(u.CustomHost); err != nil {
	// 	return fmt.Errorf("custom host '%s' is invalid: %w ", u.CustomHost, err)
	// }
	//
	// if err := validate.IsFile(u.EnvFile); err != nil {
	// 	return fmt.Errorf("the path to .env '%s' is not a file: %w", u.EnvFile, err)
	// }
	//
	// if err := validate.PathExists(u.ManifestDir); err != nil {
	// 	return fmt.Errorf("the manifest directory path '%s' is not a valid path: %w", u.ManifestDir, err)
	// }
	//
	// if u.Reset && (u.EnvFile != "" || u.ManifestDir != "") {
	// 	return fmt.Errorf("cannot specify custom files when Reset is true; Reset uses embedded defaults")
	// }

	return nil
}
