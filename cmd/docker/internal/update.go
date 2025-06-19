package internal

import (
	_ "embed"
	"epos-opensource/common"
	"fmt"
	"net/url"
)

// update logic:
// find the old env, if it does not exist give an error
// if it exists, create a copy of it in a tmp dir
// if force is set do a docker compose down on the original env
// then remove the contents of the env dir and create the updated env file and docker-compose
// if pullImages is set before deploying pull the images for the updated env
// deploy the updated compose
// if everything goes right, delete the tmp dir and finish
// else restore the tmp dir, deploy the old restored env and give an error in output
func Update(envFile, composeFile, path, name string, force, pullImages bool) (portalURL, gatewayURL string, err error) {
	common.PrintStep("Updating environment: %s", name)

	// Find the old env, if it does not exist give an error
	dir, err := common.GetEnvDir(path, name, pathPrefix)
	if err != nil {
		return "", "", fmt.Errorf("failed to get environment directory: %w", err)
	}

	common.PrintInfo("Environment directory: %s", dir)

	// If it exists, create a copy of it in a tmp dir
	tmpDir, err := createTmpCopy(dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to create backup copy: %w", err)
	}

	// Cleanup function to restore from tmp if needed
	restoreFromTmp := func() error {
		common.PrintStep("Restoring environment from backup")
		if err := common.RemoveEnvDir(dir); err != nil {
			common.PrintError("Failed to remove corrupted directory: %v", err)
		}
		if err := restoreTmpDir(tmpDir, dir); err != nil {
			return fmt.Errorf("failed to restore from backup: %w", err)
		}
		if err := deployStack(dir, name); err != nil {
			return fmt.Errorf("failed to deploy restored environment: %w", err)
		}
		return nil
	}

	common.PrintStep("Updating stack")

	// If force is set do a docker compose down on the original env
	if force {
		if err := downStack(dir, true); err != nil {
			if restoreErr := restoreFromTmp(); restoreErr != nil {
				common.PrintError("Restore failed: %v", restoreErr)
			}
			if cleanupErr := removeTmpDir(tmpDir); cleanupErr != nil {
				common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
			}
			return "", "", fmt.Errorf("docker compose down failed: %w", err)
		}
		common.PrintDone("Stopped environment: %s", name)
	}

	common.PrintStep("Removing old environment directory")

	// Remove the contents of the env dir and create the updated env file and docker-compose
	if err := common.RemoveEnvDir(dir); err != nil {
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			common.PrintError("Restore failed: %v", restoreErr)
		}
		if cleanupErr := removeTmpDir(tmpDir); cleanupErr != nil {
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
		if cleanupErr := removeTmpDir(tmpDir); cleanupErr != nil {
			common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return "", "", fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	common.PrintDone("Updated environment created in directory: %s", dir)

	// If pullImages is set before deploying pull the images for the updated env
	if pullImages {
		if err := pullEnvImages(dir, name); err != nil {
			common.PrintError("Pulling images failed: %v", err)
			if restoreErr := restoreFromTmp(); restoreErr != nil {
				common.PrintError("Restore failed: %v", restoreErr)
			}
			if cleanupErr := removeTmpDir(tmpDir); cleanupErr != nil {
				common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
			}
			return "", "", err
		}
	}

	// Deploy the updated compose
	if err := deployStack(dir, name); err != nil {
		common.PrintError("Deploy failed: %v", err)
		if restoreErr := restoreFromTmp(); restoreErr != nil {
			common.PrintError("Restore failed: %v", restoreErr)
		}
		if cleanupErr := removeTmpDir(tmpDir); cleanupErr != nil {
			common.PrintError("Failed to cleanup tmp dir: %v", cleanupErr)
		}
		return "", "", err
	}

	// If everything goes right, delete the tmp dir and finish
	if err := removeTmpDir(tmpDir); err != nil {
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
