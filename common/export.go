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

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0o750)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	filePath := filepath.Join(dir, filename)
	if err := CreateFileWithContent(filePath, string(content), false); err != nil {
		return fmt.Errorf("failed to write content to file %s: %w", filePath, err)
	}

	return nil
}
