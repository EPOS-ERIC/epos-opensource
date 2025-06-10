package internal

import (
	_ "embed"
	"epos-cli/common"
	"fmt"
	"os"
	"path"
	"runtime"
)

//go:embed static/docker-compose.yaml
var composeFile string

//go:embed static/.env
var envFile string

var configPath string

func init() {
	// get (and create if it does not exist) the dir where to put the env directories by default
	dir := "epos-cli"
	switch runtime.GOOS {
	case "windows": // If on windows, use the Appdata dir
		configPath = path.Join(os.Getenv("APPDATA"), dir)
	case "linux": // If on linux, use the ~ dir
		home, err := os.UserHomeDir()
		if err != nil {
			common.PrintError("failed to get user home directory on Linux: %v", err)
			os.Exit(1)
		}
		configPath = home
	case "darwin": // If on mac, use the '~/Library/Application Support' dir
		home, err := os.UserHomeDir()
		if err != nil {
			common.PrintError("failed to get user home directory on macOS: %v", err)
			os.Exit(1)
		}
		configPath = path.Join(home, "Library", "Application Support", dir)
	default:
		common.PrintError("unsupported operating system: %s", runtime.GOOS)
		os.Exit(1)
	}

	// Create the directory if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err = os.MkdirAll(configPath, 0700)
		if err != nil {
			common.PrintError("failed to create config directory %s: %v", configPath, err)
			os.Exit(1)
		}
	}
}

// buildEnvPath constructs the environment directory path
func buildEnvPath(customPath, name string) string {
	var basePath string
	if customPath != "" {
		basePath = customPath
	} else {
		basePath = configPath
	}
	return path.Join(basePath, name)
}

// GetEnvDir validates that the full directory path exists and returns it
func GetEnvDir(customPath, name string) (string, error) {
	envPath := buildEnvPath(customPath, name)

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return "", fmt.Errorf("directory %s does not exist: %w", envPath, err)
	} else if err != nil {
		return "", fmt.Errorf("failed to check directory %s: %w", envPath, err)
	}
	return envPath, nil
}

// NewEnvDir creates a new environment directory with .env and docker-compose.yaml files
func NewEnvDir(customEnvFile, customComposeFile, customPath, name string) (string, error) {
	envPath := buildEnvPath(customPath, name)

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

	// Create .env file
	if err := createFileWithContent(path.Join(envPath, ".env"), getContentOrDefault(customEnvFile, envFile)); err != nil {
		return "", fmt.Errorf("failed to create .env file: %w", err)
	}

	// Create docker-compose.yaml file
	if err := createFileWithContent(path.Join(envPath, "docker-compose.yaml"), getContentOrDefault(customComposeFile, composeFile)); err != nil {
		return "", fmt.Errorf("failed to create docker-compose.yaml file: %w", err)
	}

	return envPath, nil
}

// Helper function to create a file with given content
func createFileWithContent(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content to file %s: %w", filePath, err)
	}

	return nil
}

// Helper function to return custom content or default
func getContentOrDefault(custom, defaultContent string) string {
	if custom != "" {
		return custom
	}
	return defaultContent
}

func DeleteEnvDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", path)
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}
	return nil
}
