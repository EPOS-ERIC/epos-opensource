package configdir

import (
	"os"
	"path"
	"runtime"
)

var configPath string

func init() {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			panic("failed to get user home directory on macOS: " + err.Error())
		}
		configPath = path.Join(home, "Library", "Application Support", "epos-opensource")
	case "windows":
		configPath = path.Join(os.Getenv("APPDATA"), "epos-opensource")
	default: // linux and others
		home, err := os.UserHomeDir()
		if err != nil {
			panic("failed to get user home directory on Linux: " + err.Error())
		}
		configPath = path.Join(home, ".epos-opensource")
	}
}

// GetConfigPath returns the platform-specific configuration directory path
func GetConfigPath() string {
	return configPath
}
