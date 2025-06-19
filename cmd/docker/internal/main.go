package internal

import (
	_ "embed"
	"epos-opensource/common"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/joho/godotenv"
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
	if err := os.MkdirAll(envPath, 0777); err != nil {
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

func buildEnvURLs(dir string) (portalURL, gatewayURL string, err error) {
	env, err := godotenv.Read(filepath.Join(dir, ".env"))
	if err != nil {
		return "", "", fmt.Errorf("failed to read .env file at %s: %w", filepath.Join(dir, ".env"), err)
	}

	dataportalPort, ok := env["DATAPORTAL_PORT"]
	if !ok {
		return "", "", fmt.Errorf("environment variable DATAPORTAL_PORT is not set")
	}

	gatewayPort, ok := env["GATEWAY_PORT"]
	if !ok {
		return "", "", fmt.Errorf("environment variable GATEWAY_PORT is not set")
	}

	apiPath, ok := env["API_PATH"]
	if !ok {
		return "", "", fmt.Errorf("environment variable API_PATH is not set")
	}

	localIP, err := common.GetLocalIP()
	if err != nil {
		return "", "", fmt.Errorf("error getting local IP address: %w", err)
	}

	portalURL = fmt.Sprintf("http://%s:%s", localIP, dataportalPort)

	gatewayURL, err = url.JoinPath(fmt.Sprintf("http://%s:%s", localIP, gatewayPort), apiPath, "ui")
	if err != nil {
		return "", "", fmt.Errorf("error building gateway url: %w", err)
	}

	return portalURL, gatewayURL, nil
}
