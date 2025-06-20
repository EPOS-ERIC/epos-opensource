// Package internal contains the internal functions used by the kubernetes cmd to manage the environments
package internal

import (
	_ "embed"
	"epos-opensource/common"
	"fmt"
	"net/url"
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
func Update(envFile, composeFile, path, name string, force bool) (portalURL, gatewayURL string, err error) {
	common.PrintStep("Updating environment: %s", name)

	// Find the old env, if it does not exist give an error
	dir, err := common.GetEnvDir(path, name, pathPrefix)
	if err != nil {
		return "", "", fmt.Errorf("failed to get environment directory: %w", err)
	}

	common.PrintInfo("Environment directory: %s", dir)

	// If it exists, create a copy of it in a tmp dir
	tmpDir, err := common.CreateTmpCopy(dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to create backup copy: %w", err)
	}

	// Cleanup function to restore from tmp if needed
	restoreFromTmp := func() error {
		common.PrintStep("Restoring environment from backup")
		if err := common.RemoveEnvDir(dir); err != nil {
			common.PrintError("Failed to remove corrupted directory: %v", err)
		}
		if err := common.RestoreTmpDir(tmpDir, dir); err != nil {
			return fmt.Errorf("failed to restore from backup: %w", err)
		}
		if err := deployManifests(dir, name, true); err != nil {
			return fmt.Errorf("failed to deploy restored environment: %w", err)
		}
		return nil
	}

	common.PrintStep("Updating stack")

	// If force is set delete the original env
	if force {
		if err := deleteNamespace(name); err != nil {
			if restoreErr := restoreFromTmp(); restoreErr != nil {
				common.PrintError("Restore failed: %v", restoreErr)
			}
			if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
				common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
			}
			return "", "", fmt.Errorf("kubernetes namespace deletion failed: %w", err)
		}
		common.PrintDone("Stopped environment: %s", name)
	}

	common.PrintStep("Removing old environment directory")

	// Remove the contents of the env dir and create the updated env file and manifests
	if err := common.RemoveEnvDir(dir); err != nil {
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			common.PrintError("Restore failed: %v", restoreErr)
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return "", "", fmt.Errorf("failed to remove directory %s: %w", dir, err)
	}

	common.PrintStep("Creating new environment directory")

	dir, err = NewEnvDir(envFile, composeFile, path, name)
	if err != nil {
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			common.PrintError("Restore failed: %v", restoreErr)
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return "", "", fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	common.PrintDone("Updated environment created in directory: %s", dir)

	// Deploy the updated manifests
	if err := deployManifests(dir, name, force); err != nil {
		common.PrintError("Deploy failed: %v", err)
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			common.PrintError("Restore failed: %v", restoreErr)
		}
		if cleanupErr := common.RemoveTmpDir(tmpDir); cleanupErr != nil {
			common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return "", "", err
	}

	// If everything goes right, delete the tmp dir and finish
	if err := common.RemoveTmpDir(tmpDir); err != nil {
		common.PrintError("Failed to cleanup tmp dir: %v", err)
		// Don't return error as the main operation succeeded
	}

	portalURL, gatewayURL, err = buildEnvURLs(dir)
	if err != nil {
		return "", "", fmt.Errorf("error building env urls for environment '%s': %w", dir, err)
	}

	gatewayURL, err = url.JoinPath(gatewayURL, "ui/")
	return portalURL, gatewayURL, err
}
