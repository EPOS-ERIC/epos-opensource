package config

import (
	_ "embed"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var defaultConfig []byte

// GetDefaultConfig returns the default DockerEnvConfig
func GetDefaultConfig() *EnvConfig {
	var config EnvConfig
	err := yaml.Unmarshal(defaultConfig, &config)
	if err != nil {
		panic(fmt.Errorf("error unmarshalling default config yaml: %w", err))
	}
	return &config
}

func GetDefaultConfigBytes() []byte {
	return defaultConfig
}

// LoadConfig loads a DockerEnvConfig from YAML data
func LoadConfig(data []byte) (*EnvConfig, error) {
	var config EnvConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config yaml: %w", err)
	}
	return &config, nil
}

func LoadConfigFromFile(path string) (*EnvConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	config, err := LoadConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return config, nil
}

func (e *EnvConfig) BuildEnvURLs() (*common.URLs, error) {
	buildURL := func(port int, basePath string) (string, error) {
		base := &url.URL{
			Scheme: e.Protocol,
			Host:   net.JoinHostPort(e.Domain, strconv.Itoa(port)),
		}
		return url.JoinPath(base.String(), basePath)
	}

	guiURL, err := buildURL(e.Components.PlatformGUI.Port, e.Components.PlatformGUI.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("error building GUI URL: %w", err)
	}

	apiURL, err := buildURL(e.Components.Gateway.Port, e.Components.Gateway.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("error building API URL: %w", err)
	}

	urls := &common.URLs{
		GUIURL: guiURL,
		APIURL: apiURL,
	}

	if e.Components.Backoffice.Enabled {
		backofficeURL, err := buildURL(e.Components.Backoffice.GUI.Port, e.Components.Backoffice.GUI.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("error building Backoffice URL: %w", err)
		}
		urls.BackofficeURL = &backofficeURL
	}

	return urls, nil
}
