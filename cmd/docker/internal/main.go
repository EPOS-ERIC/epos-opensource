package internal

import (
	_ "embed"
	"epos-cli/common"
	"fmt"
	"os"
	"path"
	"runtime"
)

//go:embed docker-compose.yaml
var composeFile string

//go:embed .env
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
		err = os.MkdirAll(configPath, 0755)
		if err != nil {
			common.PrintError("failed to create config directory %s: %v", configPath, err)
			os.Exit(1)
		}
	}
}

// take the env file contents, a compose file contents and a path, and create a tmp dir in that path for the environment, and return the path
func NewEnvDir(customEnvFile, customComposeFile, customPath, name string) (string, error) {
	p, err := GetEnvDir(customPath, name)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		return "", fmt.Errorf("directory %s already exists", p)
	}
	err = os.MkdirAll(p, 0755) // TODO check permissions
	if err != nil {
		return "", fmt.Errorf("failed to create env directory %s: %w", p, err)
	}

	// create the .env file
	envPath := path.Join(p, ".env")
	env, err := os.Create(envPath)
	if err != nil {
		return "", fmt.Errorf("failed to create .env file at %s: %w", envPath, err)
	}
	defer env.Close()

	var envContent string
	if customEnvFile != "" {
		envContent = customEnvFile
	} else {
		envContent = envFile
	}

	if _, err := env.WriteString(envContent); err != nil {
		return "", fmt.Errorf("failed to write content to .env file at %s: %w", envPath, err)
	}

	// create the docker-compose.yaml file
	composePath := path.Join(p, "docker-compose.yaml")
	compose, err := os.Create(composePath)
	if err != nil {
		return "", fmt.Errorf("failed to create docker-compose.yaml file at %s: %w", composePath, err)
	}
	defer compose.Close()

	var composeContent string
	if customComposeFile != "" {
		composeContent = customComposeFile
	} else {
		composeContent = composeFile
	}

	if _, err := compose.WriteString(composeContent); err != nil {
		return "", fmt.Errorf("failed to write content to docker-compose.yaml file at %s: %w", composePath, err)
	}

	return p, nil
}

func GetEnvDir(customPath, name string) (string, error) {
	if customPath != "" {
		return path.Join(customPath, name), nil
	}
	return path.Join(configPath, name), nil
}

func DeleteEnvDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", path)
	}
	return os.RemoveAll(path)
}
