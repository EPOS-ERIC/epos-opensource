package config

import (
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
)

//go:embed all:helm/**
var chartFS embed.FS

// GetDefaultConfig returns a parsed copy of the embedded default K8s values configuration.
func GetDefaultConfig() *Config {
	var config Config

	err := yaml.Unmarshal(GetDefaultConfigBytes(), &config)
	if err != nil {
		panic(fmt.Errorf("error unmarshalling default config yaml: %w", err))
	}
	return &config
}

// GetDefaultConfigBytes returns the raw embedded Helm values.yaml bytes.
func GetDefaultConfigBytes() []byte {
	out, err := chartFS.ReadFile("helm/values.yaml")
	if err != nil {
		panic(fmt.Errorf("error reading values.yaml: %w", err))
	}
	return out
}

// LoadConfig loads a K8s configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config yaml: %w", err)
	}

	return &config, nil
}

// Save writes the current configuration as YAML to path.
func (e *Config) Save(path string) error {
	bytes, err := yaml.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := common.CreateFileWithContent(path, string(bytes), false); err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}

	return nil
}

// BuildEnvURLs builds GUI, API, and optional backoffice URLs from the configuration.
func (e *Config) BuildEnvURLs() (*common.URLs, error) {
	buildURL := func(paths ...string) (string, error) {
		base := &url.URL{
			Scheme: e.Protocol,
			Host:   e.Domain,
		}

		path := base.String()

		if e.URLPrefixNamespace {
			var err error
			path, err = url.JoinPath(path, e.Name)
			if err != nil {
				return "", fmt.Errorf("error joining namespace to base URL: %w", err)
			}
		}

		return url.JoinPath(path, paths...)
	}

	guiURL, err := buildURL(e.Components.PlatformGUI.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("error building GUI URL: %w", err)
	}

	apiURL, err := buildURL(e.Components.Gateway.BaseURL, "/ui")
	if err != nil {
		return nil, fmt.Errorf("error building API URL: %w", err)
	}

	urls := &common.URLs{
		GUIURL: guiURL,
		APIURL: apiURL,
	}

	if e.Components.Backoffice.Enabled {
		backofficeURL, err := buildURL(e.Components.Backoffice.GUI.BaseURL, "/home")
		if err != nil {
			return nil, fmt.Errorf("error building Backoffice URL: %w", err)
		}
		urls.BackofficeURL = &backofficeURL
	}

	return urls, nil
}

// Validate checks whether the configuration contains all required values.
func (e *Config) Validate() error {
	// Basic required fields
	if e.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if e.Protocol != "http" && e.Protocol != "https" {
		return fmt.Errorf("protocol must be http or https")
	}

	// At least one image is required
	if e.Images.RabbitmqImage == "" &&
		e.Images.DataportalImage == "" &&
		e.Images.GatewayImage == "" &&
		e.Images.MetadataDatabaseImage == "" &&
		e.Images.ResourcesServiceImage == "" &&
		e.Images.IngestorServiceImage == "" &&
		e.Images.ExternalAccessImage == "" {
		return fmt.Errorf("at least one image is required")
	}

	// Platform GUI validation
	if e.Components.PlatformGUI.BaseURL == "" {
		return fmt.Errorf("platform gui base url is required")
	}
	if !strings.HasPrefix(e.Components.PlatformGUI.BaseURL, "/") {
		return fmt.Errorf("platform gui base url must start with /")
	}
	if !strings.HasSuffix(e.Components.PlatformGUI.BaseURL, "/") {
		return fmt.Errorf("platform gui base url must end with /")
	}

	// Gateway validation
	if e.Components.Gateway.BaseURL == "" {
		return fmt.Errorf("gateway base url is required")
	}
	if !strings.HasPrefix(e.Components.Gateway.BaseURL, "/") {
		return fmt.Errorf("gateway base url must start with /")
	}
	if !strings.HasSuffix(e.Components.Gateway.BaseURL, "/api/v1") &&
		!strings.HasSuffix(e.Components.Gateway.BaseURL, "/api/v1/") {
		return fmt.Errorf("gateway base url must end with /api/v1 or /api/v1/")
	}

	// AAI validation
	if e.Components.Gateway.Aai.Enabled {
		if e.Components.Gateway.Aai.ServiceEndpoint == "" {
			return fmt.Errorf("aai service endpoint is required when aai is enabled")
		}
		if e.Components.Gateway.Aai.SecurityKey == "" {
			return fmt.Errorf("aai security key is required when aai is enabled")
		}
	}

	// Backoffice validation
	if e.Components.Backoffice.Enabled {
		if e.Components.Backoffice.GUI.BaseURL == "" {
			return fmt.Errorf("backoffice base url is required when backoffice is enabled")
		}
		if !strings.HasPrefix(e.Components.Backoffice.GUI.BaseURL, "/") {
			return fmt.Errorf("backoffice base url must start with /")
		}
		if !strings.HasSuffix(e.Components.Backoffice.GUI.BaseURL, "/") {
			return fmt.Errorf("backoffice base url must end with /")
		}
		if e.Images.BackofficeServiceImage == "" {
			return fmt.Errorf("backoffice service image is required when backoffice is enabled")
		}
		if e.Images.BackofficeUIImage == "" {
			return fmt.Errorf("backoffice ui image is required when backoffice is enabled")
		}
	}

	// Converter validation
	if e.Components.Converter.Enabled {
		if e.Images.ConverterServiceImage == "" {
			return fmt.Errorf("converter service image is required when converter is enabled")
		}
		if e.Images.ConverterRoutineImage == "" {
			return fmt.Errorf("converter routine image is required when converter is enabled")
		}
	}

	// Sharing service validation
	if e.Components.SharingService.Enabled {
		if e.Images.SharingServiceImage == "" {
			return fmt.Errorf("sharing service image is required when sharing service is enabled")
		}
	}

	// Resources service validation
	if e.Components.ResourcesService.CacheTTL <= 0 {
		return fmt.Errorf("resources service cache ttl must be greater than 0")
	}

	// RabbitMQ validation
	if e.Components.Rabbitmq.Host == "" {
		return fmt.Errorf("rabbitmq host is required")
	}
	if e.Components.Rabbitmq.Username == "" {
		return fmt.Errorf("rabbitmq username is required")
	}
	if e.Components.Rabbitmq.Password == "" {
		return fmt.Errorf("rabbitmq password is required")
	}
	if e.Components.Rabbitmq.Vhost == "" {
		return fmt.Errorf("rabbitmq vhost is required")
	}

	// Metadata database validation
	if e.Components.MetadataDatabase.User == "" {
		return fmt.Errorf("metadata database user is required")
	}
	if e.Components.MetadataDatabase.Password == "" {
		return fmt.Errorf("metadata database password is required")
	}
	if e.Components.MetadataDatabase.Host == "" {
		return fmt.Errorf("metadata database host is required")
	}
	if e.Components.MetadataDatabase.Port == 0 {
		return fmt.Errorf("metadata database port is required")
	}
	if e.Components.MetadataDatabase.Port < 1 || e.Components.MetadataDatabase.Port > 65535 {
		return fmt.Errorf("metadata database port must be between 1 and 65535")
	}
	if e.Components.MetadataDatabase.DBName == "" {
		return fmt.Errorf("metadata database db_name is required")
	}
	if e.Components.MetadataDatabase.ConnectionPoolInitSize <= 0 {
		return fmt.Errorf("metadata database connection_pool_init_size must be greater than 0")
	}
	if e.Components.MetadataDatabase.ConnectionPoolMinSize <= 0 {
		return fmt.Errorf("metadata database connection_pool_min_size must be greater than 0")
	}
	if e.Components.MetadataDatabase.ConnectionPoolMaxSize <= 0 {
		return fmt.Errorf("metadata database connection_pool_max_size must be greater than 0")
	}
	if e.Components.MetadataDatabase.ConnectionPoolInitSize > e.Components.MetadataDatabase.ConnectionPoolMinSize {
		return fmt.Errorf("metadata database connection_pool_init_size must be <= connection_pool_min_size")
	}
	if e.Components.MetadataDatabase.ConnectionPoolMinSize > e.Components.MetadataDatabase.ConnectionPoolMaxSize {
		return fmt.Errorf("metadata database connection_pool_min_size must be <= connection_pool_max_size")
	}

	// Email sender service validation
	if e.Components.EmailSenderService.Enabled {
		if e.Components.EmailSenderService.EnvironmentType != "development" &&
			e.Components.EmailSenderService.EnvironmentType != "production" &&
			e.Components.EmailSenderService.EnvironmentType != "staging" {
			return fmt.Errorf("email sender service environment_type must be development, production, or staging")
		}
		if e.Components.EmailSenderService.Sender == "" {
			return fmt.Errorf("email sender service sender is required when email sender is enabled")
		}
		if e.Components.EmailSenderService.SenderName == "" {
			return fmt.Errorf("email sender service sender_name is required when email sender is enabled")
		}
		if e.Components.EmailSenderService.MailType != "API" {
			return fmt.Errorf("email sender service mail_type must be API")
		}
		if e.Components.EmailSenderService.SenderDomain == "" {
			return fmt.Errorf("email sender service sender_domain is required when email sender is enabled")
		}
		if e.Components.EmailSenderService.MailAPIURL == "" {
			return fmt.Errorf("email sender service mail_api_url is required when email sender is enabled")
		}
		if e.Components.EmailSenderService.MailAPIKey == "" {
			return fmt.Errorf("email sender service mail_api_key is required when email sender is enabled")
		}
		if e.Images.EmailSenderServiceImage == "" {
			return fmt.Errorf("email sender service image is required when email sender is enabled")
		}
	}

	// Monitoring validation
	if e.Monitoring.Enabled {
		if e.Monitoring.URL == "" {
			return fmt.Errorf("monitoring url is required when monitoring is enabled")
		}
		if e.Monitoring.User == "" {
			return fmt.Errorf("monitoring user is required when monitoring is enabled")
		}
		if e.Monitoring.Password == "" {
			return fmt.Errorf("monitoring password is required when monitoring is enabled")
		}
	}

	return nil
}

// AsValues converts Config into Helm chart values.
func (e *Config) AsValues() (*chartutil.Values, error) {
	configYAML, err := yaml.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("marshal env config to YAML: %w", err)
	}

	values, err := chartutil.ReadValues(configYAML)
	if err != nil {
		return nil, fmt.Errorf("parse chart values from YAML: %w", err)
	}

	return &values, nil
}

// GetChart loads the embedded EPOS Helm chart from the binary filesystem.
func GetChart() (*chart.Chart, error) {
	var files []*loader.BufferedFile

	err := fs.WalkDir(chartFS, "helm", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(chartFS, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded chart file %s: %w", path, err)
		}

		relPath, err := filepath.Rel("helm", path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for embedded chart file %s: %w", path, err)
		}

		files = append(files, &loader.BufferedFile{
			Name: filepath.ToSlash(relPath),
			Data: data,
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded chart: %w", err)
	}

	c, err := loader.LoadFiles(files)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded chart: %w", err)
	}

	if c == nil {
		return nil, fmt.Errorf("chart is nil")
	}

	return c, nil
}
