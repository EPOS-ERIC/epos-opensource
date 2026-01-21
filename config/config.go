// Package config provides configuration management for epos-opensource,
// including platform-specific application data directory path and user-configurable settings.
// It determines the correct storage location based on the runtime OS:
//   - macOS: $HOME/Library/Application Support/epos-opensource
//   - Windows: %LOCALAPPDATA%/epos-opensource (falls back to %APPDATA%)
//   - Linux & others: ${XDG_DATA_HOME:-$HOME/.local/share}/epos-opensource
//
// GetPath returns the resolved directory path for storing application data.
// DefaultConfig returns the default configuration with sensible defaults.
package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

const (
	darwinOpenCommand = "open"
	linuxOpenCommand  = "xdg-open"
)

var (
	dataPath   string
	configPath string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	switch runtime.GOOS {
	case "darwin":
		dataPath = path.Join(home, "Library", "Application Support", "epos-opensource")
		configPath = path.Join(home, ".config")
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			panic("APPDATA is not set")
		}
		dataPath = path.Join(appdata, "epos-opensource")
		configPath = appdata
	default:
		dataDir := os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			dataDir = path.Join(home, ".local", "share")
		}
		dataPath = path.Join(dataDir, "epos-opensource")
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = path.Join(home, ".config")
		}
		configPath = configDir
	}
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	cfg := Config{
		TUI: TUIConfig{
			FilePickerMode: FilePickerModeNative,
		},
	}
	switch runtime.GOOS {
	case "darwin":
		cfg.TUI.OpenURLCommand = darwinOpenCommand
		cfg.TUI.OpenDirectoryCommand = darwinOpenCommand
		cfg.TUI.OpenFileCommand = darwinOpenCommand
	case "windows":
		cfg.TUI.OpenURLCommand = "cmd /c start"
		cfg.TUI.OpenDirectoryCommand = "explorer"
		cfg.TUI.OpenFileCommand = "explorer"
	default: // linux and others
		cfg.TUI.OpenURLCommand = linuxOpenCommand
		cfg.TUI.OpenDirectoryCommand = linuxOpenCommand
		cfg.TUI.OpenFileCommand = linuxOpenCommand
	}
	return cfg
}

// ValidateConfig validates the configuration
func ValidateConfig(cfg Config) error {
	if cfg.TUI.OpenURLCommand == "" {
		return errors.New("OpenURLCommand cannot be empty")
	}
	if cfg.TUI.OpenDirectoryCommand == "" {
		return errors.New("OpenDirectoryCommand cannot be empty")
	}
	if cfg.TUI.OpenFileCommand == "" {
		return errors.New("OpenFileCommand cannot be empty")
	}
	if cfg.TUI.FilePickerMode != FilePickerModeNative && cfg.TUI.FilePickerMode != FilePickerModeTUI {
		return errors.New("FilePickerMode must be 'native' or 'tui'")
	}
	return nil
}

// GetDataPath returns the platform-specific application data directory path
func GetDataPath() string {
	return dataPath
}

// GetConfigPath returns the platform-specific config file path
func GetConfigPath() string {
	return filepath.Join(configPath, "epos-opensource.yaml")
}

// LoadConfig loads the user configuration, falling back to defaults
func LoadConfig() (Config, error) {
	p := GetConfigPath()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), nil
	}
	if err := ValidateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(cfg Config) error {
	p := GetConfigPath()
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}
