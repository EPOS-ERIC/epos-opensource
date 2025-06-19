// Package common contains common functions used throughout the cli commands. Functions like Prints, CreateFileWithContent and GetEnvDir
package common

import (
	"fmt"
	"os"
	"path"
)

var configPath string

func init() {
	// Initialize the platform-specific config path
	initConfigPath()

	// Create the directory if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err = os.MkdirAll(configPath, 0700)
		if err != nil {
			PrintError("failed to create config directory %s: %v", configPath, err)
			os.Exit(1)
		}
	}
}

// CreateFileWithContent creates a file with given content
func CreateFileWithContent(filePath, content string) error {
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

// GetContentFromPathOrDefault reads content from filePath if provided, otherwise returns defaultContent
func GetContentFromPathOrDefault(filePath, defaultContent string) (string, error) {
	if filePath == "" {
		return defaultContent, nil
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
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

// BuildEnvPath constructs the environment directory path
func BuildEnvPath(customPath, name, prefix string) string {
	var basePath string
	if customPath != "" {
		basePath = customPath
	} else {
		basePath = path.Join(configPath, prefix)
	}
	return path.Join(basePath, name)
}

// GetEnvDir validates that the full directory path exists and returns it
func GetEnvDir(customPath, name, prefix string) (string, error) {
	envPath := BuildEnvPath(customPath, name, prefix)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return "", fmt.Errorf("directory %s does not exist: %w", envPath, err)
	} else if err != nil {
		return "", fmt.Errorf("failed to check directory %s: %w", envPath, err)
	}
	return envPath, nil
}

// RemoveEnvDir deletes the environment directory with logs
func RemoveEnvDir(dir string) error {
	PrintStep("Deleting environment directory: %s", dir)
	if err := DeleteEnvDir(dir); err != nil {
		return err
	}
	PrintDone("Deleted environment directory: %s", dir)
	return nil
}
