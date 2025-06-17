package internal

import (
	_ "embed"
	"epos-cli/common"
	"fmt"
	"os"
	"path"
)

const pathPrefix = "docker"

//go:embed static/docker-compose.yaml
var composeFile string

//go:embed static/.env
var envFile string

// NewEnvDir creates a new environment directory with .env and docker-compose.yaml files.
// If customEnvFilePath or customComposeFilePath are provided, it reads the content from those files.
// If they are empty strings, it uses default content for the respective files.
// Returns the path to the created environment directory.
// If any error occurs after directory creation, the directory and its contents are automatically cleaned up.
func NewEnvDir(customEnvFilePath, customComposeFilePath, customPath, name string) (string, error) {
	envPath := common.BuildEnvPath(customPath, name, pathPrefix)

	// Check if directory already exists
	if _, err := os.Stat(envPath); err == nil {
		return "", fmt.Errorf("directory %s already exists", envPath)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check directory %s: %w", envPath, err)
	}

	// Create the directory
	if err := os.MkdirAll(envPath, 0700); err != nil {
		return "", fmt.Errorf("failed to create env directory %s: %w", envPath, err)
	}

	var success bool
	// Ensure cleanup of directory if any error occurs after creation
	defer func() {
		if !success {
			if removeErr := os.RemoveAll(envPath); removeErr != nil {
				common.PrintError("Failed to cleanup directory '%s' after error: %v. You may need to remove it manually.", envPath, removeErr)
			}
		}
	}()

	// Get .env file content (from file path or use default)
	envContent, err := common.GetContentFromPathOrDefault(customEnvFilePath, envFile)
	if err != nil {
		return "", fmt.Errorf("failed to get .env file content: %w", err)
	}

	// Create .env file
	if err := common.CreateFileWithContent(path.Join(envPath, ".env"), envContent); err != nil {
		return "", fmt.Errorf("failed to create .env file: %w", err)
	}

	// Get docker-compose.yaml file content (from file path or use default)
	composeContent, err := common.GetContentFromPathOrDefault(customComposeFilePath, composeFile)
	if err != nil {
		return "", fmt.Errorf("failed to get docker-compose.yaml file content: %w", err)
	}

	// Create docker-compose.yaml file
	if err := common.CreateFileWithContent(path.Join(envPath, "docker-compose.yaml"), composeContent); err != nil {
		return "", fmt.Errorf("failed to create docker-compose.yaml file: %w", err)
	}

	success = true
	return envPath, nil
}
