// Package k8score contains the internal functions used by the kubernetes cmd to manage the environments
package k8score

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
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
func Update(envFile, composeFile, name string, force bool) (portalURL, gatewayURL, backofficeURL string, err error) {
	common.PrintStep("Updating environment: %s", name)

	env, err := db.GetKubernetesByName(name)
	if err != nil {
		return "", "", "", fmt.Errorf("error getting kubernetes environment from db called '%s': %w", name, err)
	}
	portalURL = env.GuiUrl
	gatewayURL = env.ApiUrl
	backofficeURL = env.BackofficeUrl

	common.PrintInfo("Environment directory: %s", env.Directory)

	// If it exists, create a copy of it in a tmp dir
	tmpDir, err := common.CreateTmpCopy(env.Directory)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create backup copy: %w", err)
	}

	// Cleanup function to restore from tmp if needed
	restoreFromTmp := func() error {
		common.PrintStep("Restoring environment from backup")
		if err := common.RemoveEnvDir(env.Directory); err != nil {
			common.PrintError("Failed to remove corrupted directory: %v", err)
		}
		if err := common.RestoreTmpDir(tmpDir, env.Directory); err != nil {
			return fmt.Errorf("failed to restore from backup: %w", err)
		}
		if err := deployManifests(env.Directory, name, true, env.Context, env.Protocol); err != nil {
			return fmt.Errorf("failed to deploy restored environment: %w", err)
		}
		return nil
	}

	common.PrintStep("Updating stack")

	// If force is set delete the original env
	if force {
		if err := deleteNamespace(name, env.Context); err != nil {
			if restoreErr := restoreFromTmp(); restoreErr != nil {
				common.PrintError("Restore failed: %v", restoreErr)
			}
			if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
				common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
			}
			return "", "", "", fmt.Errorf("kubernetes namespace deletion failed: %w", err)
		}
		common.PrintDone("Stopped environment: %s", name)
	}

	common.PrintStep("Removing old environment directory")

	// Remove the contents of the env dir and create the updated env file and manifests
	if err := common.RemoveEnvDir(env.Directory); err != nil {
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			common.PrintError("Restore failed: %v", restoreErr)
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return "", "", "", fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
	}

	common.PrintStep("Creating new environment directory")

	dir, err := NewEnvDir(envFile, composeFile, env.Directory, name, env.Context, env.Protocol)
	if err != nil {
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			common.PrintError("Restore failed: %v", restoreErr)
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return "", "", "", fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	common.PrintDone("Updated environment created in directory: %s", dir)

	// Deploy the updated manifests
	if err := deployManifests(dir, name, force, env.Context, env.Protocol); err != nil {
		common.PrintError("Deploy failed: %v", err)
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			common.PrintError("Restore failed: %v", restoreErr)
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return "", "", "", err
	}

	// If everything goes right, delete the tmp dir and finish
	if err := common.RemoveTmpDir(tmpDir); err != nil {
		common.PrintError("Failed to cleanup tmp dir: %v", err)
		// Don't return error as the main operation succeeded
	}

	gatewayURL, err = url.JoinPath(gatewayURL, "ui/")
	return portalURL, gatewayURL, backofficeURL, err
}
