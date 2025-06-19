//go:build windows

package common

import (
	"os"
	"path"
)

func initConfigPath() {
	dir := "epos-opensource"
	configPath = path.Join(os.Getenv("APPDATA"), dir)
}
