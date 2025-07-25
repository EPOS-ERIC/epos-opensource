// Package configdir provides the platform-specific application data directory path for epos-opensource.
// It determines the correct storage location based on the runtime OS:
//   - macOS: $HOME/Library/Application Support/epos-opensource
//   - Windows: %LOCALAPPDATA%/epos-opensource (falls back to %APPDATA%)
//   - Linux & others: ${XDG_DATA_HOME:-$HOME/.local/share}/epos-opensource
//
// GetDataPath returns the resolved directory path for storing application data.
package configdir

import (
	"os"
	"path"
	"runtime"
)

var dataPath string

func init() {
	switch runtime.GOOS {
	case "darwin":
		home := must(os.UserHomeDir())
		dataPath = path.Join(home, "Library", "Application Support", "epos-opensource")
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			// fallback
			roaming := os.Getenv("APPDATA")
			if roaming == "" {
				panic("neither LOCALAPPDATA nor APPDATA is set")
			}
			local = roaming
		}
		dataPath = path.Join(local, "epos-opensource")
	default:
		home := must(os.UserHomeDir())
		dataDir := os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			dataDir = path.Join(home, ".local", "share")
		}
		dataPath = path.Join(dataDir, "epos-opensource")
	}
}

// GetPath returns the platform-specific application‑data directory path
func GetPath() string {
	return dataPath
}

func must(dir string, err error) string {
	if err != nil {
		panic(err)
	}
	return dir
}
