package config

// Config represents the application configuration
type Config struct {
	TUI TUIConfig `yaml:"tui"`
}

// TUIConfig holds TUI-specific configurations
type TUIConfig struct {
	DefaultScreen        string `yaml:"defaultScreen"`
	OpenUrlCommand       string `yaml:"openUrlCommand"`
	OpenDirectoryCommand string `yaml:"openDirectoryCommand"`
	OpenFileCommand      string `yaml:"openFileCommand"`
}
