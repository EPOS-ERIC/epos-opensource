// Package common contains common functions used throughout the cli commands. Functions like Prints, CreateFileWithContent and GetEnvDir
package common

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common/configdir"
	"github.com/epos-eu/epos-opensource/db"
)

func init() {
	// Create the directory if it doesn't exist
	if _, err := os.Stat(configdir.GetConfigPath()); os.IsNotExist(err) {
		err = os.MkdirAll(configdir.GetConfigPath(), 0777)
		if err != nil {
			PrintError("failed to create config directory %s: %v", configdir.GetConfigPath(), err)
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
		basePath = path.Join(configdir.GetConfigPath(), prefix)
	}
	return path.Join(basePath, name)
}

// GetEnvDir validates that the full directory path exists and returns it
func GetEnvDir(customPath, name, platform string) (string, error) {
	env, dbErr := db.GetEnv(name, platform)
	if dbErr != nil {
		return "", fmt.Errorf("failed to check environment in db: %w", dbErr)
	}
	if env == nil {
		return "", fmt.Errorf("environment '%s' for platform '%s' does not exist in the database", name, platform)
	}
	return env.Directory, nil
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

// CreateTmpCopy creates a backup copy of the environment directory in a temporary location
func CreateTmpCopy(dir string) (string, error) {
	PrintStep("Creating backup copy of environment")

	tmpDir, err := os.MkdirTemp("", "env-backup-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	if err := CopyDir(dir, tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to copy environment to backup: %w", err)
	}

	PrintDone("Backup created at: %s", tmpDir)
	return tmpDir, nil
}

// RestoreTmpDir restores the environment from temporary backup to target directory
func RestoreTmpDir(tmpDir, targetDir string) error {
	PrintStep("Restoring environment from backup")

	if err := os.MkdirAll(targetDir, 0777); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	if err := CopyDir(tmpDir, targetDir); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	PrintDone("Environment restored from backup")
	return nil
}

// CopyDir recursively copies a directory from src to dst
func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory %s: %w", src, err)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy subdirectory %s to %s: %w", srcPath, dstPath, err)
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s to %s: %w", srcPath, dstPath, err)
			}
		}
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file %s: %w", src, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy data from %s to %s: %w", src, dst, err)
	}

	return nil
}

// RemoveTmpDir removes the temporary backup directory with logs
func RemoveTmpDir(tmpDir string) error {
	PrintStep("Cleaning up backup directory: %s", tmpDir)

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to cleanup backup directory: %w", err)
	}

	PrintDone("Backup directory cleaned up: %s", tmpDir)
	return nil
}
