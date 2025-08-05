package dockercore

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/joho/godotenv"
)

const platform = "docker"

//go:embed static/docker-compose.yaml
var ComposeFile string

//go:embed static/.env
var EnvFile string

// NewEnvDir creates a new environment directory with .env and docker-compose.yaml files.
// If customEnvFilePath or customComposeFilePath are provided, it reads the content from those files.
// If they are empty strings, it uses default content for the respective files.
// Returns the path to the created environment directory.
// If any error occurs after directory creation, the directory and its contents are automatically cleaned up.
func NewEnvDir(customEnvFilePath, customComposeFilePath, customPath, name string) (string, error) {
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
	if err := os.MkdirAll(envPath, 0750); err != nil {
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

	// Get .env file content (from file path or use default)
	envContent, err := common.GetContentFromPathOrDefault(customEnvFilePath, EnvFile)
	if err != nil {
		return "", fmt.Errorf("failed to get .env file content: %w", err)
	}

	// Create .env file
	if err := common.CreateFileWithContent(path.Join(envPath, ".env"), envContent); err != nil {
		return "", fmt.Errorf("failed to create .env file: %w", err)
	}

	// Get docker-compose.yaml file content (from file path or use default)
	composeContent, err := common.GetContentFromPathOrDefault(customComposeFilePath, ComposeFile)
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

func buildEnvURLs(dir string, ports *DeploymentPorts, customIP string) (urls Urls, err error) {
	env, err := godotenv.Read(filepath.Join(dir, ".env"))
	if err != nil {
		return urls, fmt.Errorf("failed to read .env file at %s: %w", filepath.Join(dir, ".env"), err)
	}

	apiPath, ok := env["API_PATH"]
	if !ok {
		return urls, fmt.Errorf("environment variable API_PATH is not set")
	}

	ip := "localhost"
	if customIP != "" {
		ip = customIP
	}

	urls.guiURL = fmt.Sprintf("http://%s:%d", ip, ports.GUI)

	urls.apiURL, err = url.JoinPath(fmt.Sprintf("http://%s:%d", ip, ports.API), apiPath)
	if err != nil {
		return urls, fmt.Errorf("error building gateway URL: %w", err)
	}

	urls.backofficeURL = fmt.Sprintf("http://%s:%d", ip, ports.Backoffice)

	return urls, nil
}
