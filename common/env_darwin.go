//go:build darwin

package common

import (
	"os"
	"path"
)

func initConfigPath() {
	dir := "epos-opensource"
	home, err := os.UserHomeDir()
	if err != nil {
		PrintError("failed to get user home directory on macOS: %v", err)
		os.Exit(1)
	}
	configPath = path.Join(home, "Library", "Application Support", dir)
}
