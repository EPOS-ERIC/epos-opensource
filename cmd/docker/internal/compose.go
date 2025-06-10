package internal

import (
	"epos-cli/common"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// composeCommand creates a docker compose command configured with the given directory and environment name
func composeCommand(dir, name string, args ...string) *exec.Cmd {
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Dir = dir
	if name != "" {
		cmd.Env = append(cmd.Env, "ENV_NAME="+name)
	}
	return cmd
}

// pullEnvImages pulls docker images for the environment with custom messages
func pullEnvImages(dir, name string) error {
	common.PrintStep("Pulling images for environment %s", name)
	if err := common.RunCommand(composeCommand(dir, "", "pull")); err != nil {
		return fmt.Errorf("pull images failed: %w", err)
	}
	common.PrintDone("Images pulled for environment %s", name)
	return nil
}

// deployStack deploys the stack in the specified directory
func deployStack(dir, name string) error {
	common.PrintStep("Deploying stack")
	if err := common.RunCommand(composeCommand(dir, name, "up", "-d")); err != nil {
		return fmt.Errorf("deploy stack failed: %w", err)
	}
	common.PrintDone("Deployed environment: %s", name)
	return nil
}

// downStack stops the stack running in the given directory
func downStack(dir string) error {
	return common.RunCommand(composeCommand(dir, "", "down"))
}

// removeEnvDir deletes the environment directory with logs
func removeEnvDir(dir string) error {
	common.PrintStep("Deleting environment directory: %s", dir)
	if err := DeleteEnvDir(dir); err != nil {
		return err
	}
	common.PrintDone("Deleted environment directory: %s", dir)
	return nil
}

// deployMetadataCache deploys an nginx docker container running a file server exposing a volume
func deployMetadataCache(dir, envName string) error {
	cmd := exec.Command(
		"docker",
		"run",
		"-d",
		"--name",
		envName+"-metadata-cache",
		"-p",
		"8080:80", // TODO use a free port
		"-v",
		dir+":/usr/share/nginx/html",
		"nginx",
	)

	err := common.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("error deploying metadata-cache: %w", err)
	}

	return nil
}

// createTmpCopy creates a backup copy of the environment directory in a temporary location
func createTmpCopy(dir string) (string, error) {
	common.PrintStep("Creating backup copy of environment")

	tmpDir, err := os.MkdirTemp("", "env-backup-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	if err := copyDir(dir, tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to copy environment to backup: %w", err)
	}

	common.PrintDone("Backup created at: %s", tmpDir)
	return tmpDir, nil
}

// restoreTmpDir restores the environment from temporary backup to target directory
func restoreTmpDir(tmpDir, targetDir string) error {
	common.PrintStep("Restoring environment from backup")

	if err := os.MkdirAll(targetDir, 0700); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	if err := copyDir(tmpDir, targetDir); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	common.PrintDone("Environment restored from backup")
	return nil
}

// copyDir recursively copies a directory from src to dst
func copyDir(src, dst string) error {
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
			if err := copyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy subdirectory %s to %s: %w", srcPath, dstPath, err)
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s to %s: %w", srcPath, dstPath, err)
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
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

// removeTmpDir removes the temporary backup directory with logs
func removeTmpDir(tmpDir string) error {
	common.PrintStep("Cleaning up backup directory: %s", tmpDir)

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to cleanup backup directory: %w", err)
	}

	common.PrintDone("Backup directory cleaned up: %s", tmpDir)
	return nil
}
