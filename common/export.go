package common

import (
	"fmt"
	"os"
	"path/filepath"
)

func Export(path, filename string, content []byte) error {
	var dir string

	// If path is empty, use current directory
	if path == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		dir = currentDir
	} else {
		dir = path
	}

	// Check if directory exists, create if it doesn't
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0750)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create the file in the directory
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		return fmt.Errorf("failed to write content to file %s: %w", filePath, err)
	}

	return nil
}
