//go:build windows

package common

import (
	"os"
	"path"
)

func initConfigPath() {
	dir := "epos-cli"
	configPath = path.Join(os.Getenv("APPDATA"), dir)
}
