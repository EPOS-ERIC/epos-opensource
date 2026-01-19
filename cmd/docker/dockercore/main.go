package dockercore

import (
	_ "embed"
	"fmt"
	"os"
	"path"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore/config"
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
)

const platform = "docker"

// NewEnvDir creates a new environment directory with .env and docker-compose.yaml files.
// TODO: update documentation
// If customEnvFilePath or customComposeFilePath are provided, it reads the content from those files.
// If they are empty strings, it uses default content for the respective files.
// Returns the path to the created environment directory.
// If any error occurs after directory creation, the directory and its contents are automatically cleaned up.
func NewEnvDir(customEnvFilePath, customComposeFilePath, customPath, name string, cfg *config.EnvConfig) (string, error) {
	envPath, err := common.BuildEnvPath(customPath, name, platform)
	if err != nil {
		return "", err
	}

	// Check if directory already exists
	if _, err := os.Stat(envPath); err == nil {
		return "", fmt.Errorf("directory %s already exists", envPath)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check directory %s: %w", envPath, err)
	}

	// Create the directory
	if err := os.MkdirAll(envPath, 0o750); err != nil {
		return "", fmt.Errorf("failed to create env directory %s: %w", envPath, err)
	}

	var success bool
	// Ensure cleanup of directory if any error occurs after creation
	defer func() {
		if !success {
			if removeErr := os.RemoveAll(envPath); removeErr != nil {
				display.Error("Failed to cleanup directory '%s' after error: %v. You may need to remove it manually.", envPath, removeErr)
			}
		}
	}()

	cfg.Name = name

	files, err := cfg.Render()
	if err != nil {
		return "", fmt.Errorf("failed to render config: %w", err)
	}

	// Get .env file content (from file path or use default)
	envContent, err := common.GetContentFromPathOrDefault(customEnvFilePath, files[".env"])
	if err != nil {
		return "", fmt.Errorf("failed to get .env file content: %w", err)
	}

	// Create .env file
	if err := common.CreateFileWithContent(path.Join(envPath, ".env"), envContent); err != nil {
		return "", fmt.Errorf("failed to create .env file: %w", err)
	}

	// Get docker-compose.yaml file content (from file path or use default)
	composeContent, err := common.GetContentFromPathOrDefault(customComposeFilePath, files["docker-compose.yaml"])
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
