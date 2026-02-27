// Package common contains common functions used throughout the cli commands. Functions like display.s, CreateFileWithContent and GetEnvDir
package common

import (
	"fmt"
	"os"

	"github.com/EPOS-ERIC/epos-opensource/config"
	"github.com/EPOS-ERIC/epos-opensource/display"
)

func init() {
	// Create the directory if it doesn't exist
	if _, err := os.Stat(config.GetDataPath()); os.IsNotExist(err) {
		err = os.MkdirAll(config.GetDataPath(), 0o750)
		if err != nil {
			display.Error("failed to create config directory %s: %v", config.GetDataPath(), err)
			os.Exit(1)
		}
	}
}

// CreateFileWithContent creates a file with given content
func CreateFileWithContent(filePath, content string, doNotEdit bool) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()
	if _, err := file.WriteString(GenerateExportHeader(doNotEdit) + content); err != nil {
		return fmt.Errorf("failed to write content to file %s: %w", filePath, err)
	}
	return nil
}
