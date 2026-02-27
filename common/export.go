package common

import (
	"fmt"
	"os"
	"path/filepath"
)

// Export writes content to a file in the given directory, creating the
// directory if it does not already exist.
//
// It returns the absolute path to the exported file.
func Export(path, filename string, content []byte) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for export path %s: %w", path, err)
	}

	var dir string

	// If path is empty, use current directory
	if path == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		dir = currentDir
	} else {
		dir = path
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0o750)
		if err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	filePath := filepath.Join(dir, filename)
	if err := CreateFileWithContent(filePath, string(content), false); err != nil {
		return "", fmt.Errorf("failed to write content to file %s: %w", filePath, err)
	}

	return filepath.Join(path, filename), nil
}
