//go:build linux

package common

import (
	"os"
	"path"
)

func initConfigPath() {
	home, err := os.UserHomeDir()
	if err != nil {
		PrintError("failed to get user home directory on Linux: %v", err)
		os.Exit(1)
	}
	configPath = path.Join(home, "epos-opensource")
}
