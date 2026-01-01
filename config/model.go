package config

// Config represents the application configuration
type Config struct {
	TUI TUIConfig `yaml:"tui"`
}

// FilePickerMode represents the mode for file picker selection
type FilePickerMode string

const (
	FilePickerModeNative FilePickerMode = "native"
	FilePickerModeTUI    FilePickerMode = "tui"
)

// TUIConfig holds TUI-specific configurations
type TUIConfig struct {
	OpenURLCommand       string         `yaml:"openURLCommand"`
	OpenDirectoryCommand string         `yaml:"openDirectoryCommand"`
	OpenFileCommand      string         `yaml:"openFileCommand"`
	FilePickerMode       FilePickerMode `yaml:"filePickerMode"`
}
