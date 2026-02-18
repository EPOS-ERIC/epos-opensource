package docker

import (
	_ "embed"
	"fmt"
	"os"
	"path"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

const platform = "docker"

// NewEnvDir creates a new environment directory with config.yaml and rendered template files (.env, docker-compose.yaml).
// The customPath parameter specifies a custom base path for the environment directory.
// The cfg parameter contains the environment configuration used for templating.
// Returns the path to the created environment directory.
// If any error occurs after directory creation, the directory and its contents are automatically cleaned up.
func NewEnvDir(customPath string, cfg *config.EnvConfig) (string, error) {
	envPath, err := common.BuildEnvPath(customPath, cfg.Name, platform)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(envPath); err == nil {
		return "", fmt.Errorf("directory %s already exists", envPath)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check directory %s: %w", envPath, err)
	}

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

	err = cfg.Save(path.Join(envPath, "config.yaml"))
	if err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	files, err := cfg.Render()
	if err != nil {
		return "", fmt.Errorf("failed to render config: %w", err)
	}

	for name, content := range files {
		if err := common.CreateFileWithContent(path.Join(envPath, name), content, true); err != nil {
			return "", fmt.Errorf("failed to create .env file: %w", err)
		}
	}

	success = true
	return envPath, nil
}
