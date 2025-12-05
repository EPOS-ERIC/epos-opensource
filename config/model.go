package config

// Config represents the application configuration
type Config struct {
	Keymaps KeymapsConfig `yaml:"keymaps"`
	TUI     TUIConfig     `yaml:"tui"`
}

// KeymapsConfig holds key binding configurations
type KeymapsConfig struct {
	Up     []string `yaml:"up"`
	Down   []string `yaml:"down"`
	Select []string `yaml:"select"`
	Quit   []string `yaml:"quit"`
}

// TUIConfig holds TUI-specific configurations
type TUIConfig struct {
	DefaultScreen string `yaml:"defaultScreen"`
}
