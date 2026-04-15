// Package config provides Kubernetes/Helm configuration loading, modeling, and rendering.
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
func (c *Config) Save(path string) error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := common.CreateFileWithContent(path, string(bytes), false); err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}

	return nil
}

// BuildEnvURLs builds GUI, API, and optional backoffice URLs from the configuration.
func (c *Config) BuildEnvURLs() (*common.URLs, error) {
	buildURL := func(paths ...string) (string, error) {
		base := &url.URL{
			Scheme: c.Protocol,
			Host:   c.Domain,
		}

		path := base.String()

		if c.URLPrefixNamespace {
			var err error
			path, err = url.JoinPath(path, c.Name)
			if err != nil {
				return "", fmt.Errorf("error joining namespace to base URL: %w", err)
			}
		}

		return url.JoinPath(path, paths...)
	}

	guiURL, err := buildURL(c.Components.PlatformGUI.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("error building GUI URL: %w", err)
	}

	apiURL, err := buildURL(c.Components.Gateway.BaseURL, "/ui")
	if err != nil {
		return nil, fmt.Errorf("error building API URL: %w", err)
	}

	urls := &common.URLs{
		GUIURL: guiURL,
		APIURL: apiURL,
	}

	if c.Components.Backoffice.Enabled {
		backofficeURL, err := buildURL(c.Components.Backoffice.GUI.BaseURL, "/home")
		if err != nil {
			return nil, fmt.Errorf("error building Backoffice URL: %w", err)
		}
		urls.BackofficeURL = &backofficeURL
	}

	return urls, nil
}

func validateAuthDependsOnGatewayAAI(serviceName string, auth Auth, gatewayAAIEnabled bool) error {
	if auth.Enabled && !gatewayAAIEnabled {
		return fmt.Errorf("%s auth requires gateway aai to be enabled", serviceName)
	}

	return nil
}

func validateIngressTLSSecret(serviceName string, tls TLS) error {
	if tls.Enabled && tls.SecretName == "" {
		return fmt.Errorf("%s tls secret_name is required when tls is enabled", serviceName)
	}

	return nil
}

// Validate checks whether the configuration contains all required values.
func (c *Config) Validate() error {
	// Basic required fields
	if c.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if c.Protocol != "http" && c.Protocol != "https" {
		return fmt.Errorf("protocol must be http or https")
	}
	if err := validateIngressTLSSecret("platform gui", c.Components.PlatformGUI.TLS); err != nil {
		return err
	}
	if err := validateIngressTLSSecret("gateway", c.Components.Gateway.TLS); err != nil {
		return err
	}
	if c.Components.Backoffice.Enabled {
		if err := validateIngressTLSSecret("backoffice", c.Components.Backoffice.TLS); err != nil {
			return err
		}
	}
	if c.Components.AAIService.Enabled {
		if err := validateIngressTLSSecret("aai service", c.Components.AAIService.TLS); err != nil {
			return err
		}
	}

	// Required images
	if c.Images.DataportalImage == "" {
		return fmt.Errorf("dataportal image is required")
	}
	if c.Images.GatewayImage == "" {
		return fmt.Errorf("gateway image is required")
	}
	if c.Images.ResourcesServiceImage == "" {
		return fmt.Errorf("resources service image is required")
	}
	if c.Images.IngestorServiceImage == "" {
		return fmt.Errorf("ingestor service image is required")
	}
	if c.Images.ExternalAccessImage == "" {
		return fmt.Errorf("external access image is required")
	}
	if c.Components.Rabbitmq.Enabled && c.Images.RabbitmqImage == "" {
		return fmt.Errorf("rabbitmq image is required when rabbitmq is enabled")
	}
	if c.Components.MetadataDatabase.Enabled && c.Images.MetadataDatabaseImage == "" {
		return fmt.Errorf("metadata database image is required when metadata database is enabled")
	}
	if c.Components.AAIService.Enabled && c.Images.AAIServiceImage == "" {
		return fmt.Errorf("aai service image is required when aai service is enabled")
	}

	// Platform GUI validation
	if c.Components.PlatformGUI.BaseURL == "" {
		return fmt.Errorf("platform gui base url is required")
	}
	if !strings.HasPrefix(c.Components.PlatformGUI.BaseURL, "/") {
		return fmt.Errorf("platform gui base url must start with /")
	}
	if !strings.HasSuffix(c.Components.PlatformGUI.BaseURL, "/") {
		return fmt.Errorf("platform gui base url must end with /")
	}

	// Gateway validation
	if c.Components.Gateway.BaseURL == "" {
		return fmt.Errorf("gateway base url is required")
	}
	if !strings.HasPrefix(c.Components.Gateway.BaseURL, "/") {
		return fmt.Errorf("gateway base url must start with /")
	}
	if !strings.HasSuffix(c.Components.Gateway.BaseURL, "/api/v1") &&
		!strings.HasSuffix(c.Components.Gateway.BaseURL, "/api/v1/") {
		return fmt.Errorf("gateway base url must end with /api/v1 or /api/v1/")
	}

	// AAI validation
	if c.Components.AAIService.Enabled {
		if !c.Components.Gateway.AAI.Enabled {
			return fmt.Errorf("aai service is enabled but aai in the gateway is disabled")
		}
		if c.Components.AAIService.Name == "" {
			return fmt.Errorf("aai service name is required when aai service is enabled")
		}
		if c.Components.AAIService.Surname == "" {
			return fmt.Errorf("aai service surname is required when aai service is enabled")
		}
		if c.Components.AAIService.Email == "" {
			return fmt.Errorf("aai service email is required when aai service is enabled")
		}
		if c.Components.AAIService.Password == "" {
			return fmt.Errorf("aai service password is required when aai service is enabled")
		}
	}
	if c.Components.Gateway.AAI.Enabled {
		if c.Components.Gateway.AAI.ServiceEndpoint == "" {
			return fmt.Errorf("aai service endpoint is required when aai is enabled")
		}
	}

	authValidations := []struct {
		serviceName string
		auth        Auth
	}{
		{serviceName: "backoffice service", auth: c.Components.Backoffice.Service.Auth},
		{serviceName: "converter service", auth: c.Components.Converter.Auth},
		{serviceName: "resources service", auth: c.Components.ResourcesService.Auth},
		{serviceName: "ingestor service", auth: c.Components.IngestorService.Auth},
		{serviceName: "external access service", auth: c.Components.ExternalAccessService.Auth},
		{serviceName: "sharing service", auth: c.Components.SharingService.Auth},
		{serviceName: "email sender service", auth: c.Components.EmailSenderService.Auth},
	}
	for _, validation := range authValidations {
		if err := validateAuthDependsOnGatewayAAI(validation.serviceName, validation.auth, c.Components.Gateway.AAI.Enabled); err != nil {
			return err
		}
	}

	// Backoffice validation
	if c.Components.Backoffice.Enabled {
		if !c.Components.Backoffice.Service.Auth.Enabled {
			return fmt.Errorf("backoffice service auth must be enabled when backoffice is enabled")
		}
		if c.Components.Backoffice.GUI.BaseURL == "" {
			return fmt.Errorf("backoffice base url is required when backoffice is enabled")
		}
		if !strings.HasPrefix(c.Components.Backoffice.GUI.BaseURL, "/") {
			return fmt.Errorf("backoffice base url must start with /")
		}
		if !strings.HasSuffix(c.Components.Backoffice.GUI.BaseURL, "/") {
			return fmt.Errorf("backoffice base url must end with /")
		}
		if c.Images.BackofficeServiceImage == "" {
			return fmt.Errorf("backoffice service image is required when backoffice is enabled")
		}
		if c.Images.BackofficeUIImage == "" {
			return fmt.Errorf("backoffice ui image is required when backoffice is enabled")
		}
	}

	// Converter validation
	if c.Components.Converter.Enabled {
		if c.Images.ConverterServiceImage == "" {
			return fmt.Errorf("converter service image is required when converter is enabled")
		}
		if c.Images.ConverterRoutineImage == "" {
			return fmt.Errorf("converter routine image is required when converter is enabled")
		}
	}

	// Sharing service validation
	if c.Components.SharingService.Enabled {
		if c.Images.SharingServiceImage == "" {
			return fmt.Errorf("sharing service image is required when sharing service is enabled")
		}
	}

	// Resources service validation
	if c.Components.ResourcesService.CacheTTL <= 0 {
		return fmt.Errorf("resources service cache ttl must be greater than 0")
	}

	// RabbitMQ validation
	if c.Components.Rabbitmq.Host == "" {
		return fmt.Errorf("rabbitmq host is required")
	}
	if c.Components.Rabbitmq.Username == "" {
		return fmt.Errorf("rabbitmq username is required")
	}
	if c.Components.Rabbitmq.Password == "" {
		return fmt.Errorf("rabbitmq password is required")
	}
	if c.Components.Rabbitmq.Vhost == "" {
		return fmt.Errorf("rabbitmq vhost is required")
	}

	// Metadata database validation
	if c.Components.MetadataDatabase.User == "" {
		return fmt.Errorf("metadata database user is required")
	}
	if c.Components.MetadataDatabase.Password == "" {
		return fmt.Errorf("metadata database password is required")
	}
	if c.Components.MetadataDatabase.Host == "" {
		return fmt.Errorf("metadata database host is required")
	}
	if c.Components.MetadataDatabase.Port == 0 {
		return fmt.Errorf("metadata database port is required")
	}
	if c.Components.MetadataDatabase.Port < 1 || c.Components.MetadataDatabase.Port > 65535 {
		return fmt.Errorf("metadata database port must be between 1 and 65535")
	}
	if c.Components.MetadataDatabase.DBName == "" {
		return fmt.Errorf("metadata database db_name is required")
	}
	if c.Components.MetadataDatabase.ConnectionPoolInitSize <= 0 {
		return fmt.Errorf("metadata database connection_pool_init_size must be greater than 0")
	}
	if c.Components.MetadataDatabase.ConnectionPoolMinSize <= 0 {
		return fmt.Errorf("metadata database connection_pool_min_size must be greater than 0")
	}
	if c.Components.MetadataDatabase.ConnectionPoolMaxSize <= 0 {
		return fmt.Errorf("metadata database connection_pool_max_size must be greater than 0")
	}
	if c.Components.MetadataDatabase.ConnectionPoolInitSize > c.Components.MetadataDatabase.ConnectionPoolMinSize {
		return fmt.Errorf("metadata database connection_pool_init_size must be <= connection_pool_min_size")
	}
	if c.Components.MetadataDatabase.ConnectionPoolMinSize > c.Components.MetadataDatabase.ConnectionPoolMaxSize {
		return fmt.Errorf("metadata database connection_pool_min_size must be <= connection_pool_max_size")
	}

	// Email sender service validation
	if c.Components.EmailSenderService.Enabled {
		if c.Components.EmailSenderService.EnvironmentType != "development" &&
			c.Components.EmailSenderService.EnvironmentType != "production" &&
			c.Components.EmailSenderService.EnvironmentType != "staging" {
			return fmt.Errorf("email sender service environment_type must be development, production, or staging")
		}
		if c.Components.EmailSenderService.Sender == "" {
			return fmt.Errorf("email sender service sender is required when email sender is enabled")
		}
		if c.Components.EmailSenderService.SenderName == "" {
			return fmt.Errorf("email sender service sender_name is required when email sender is enabled")
		}
		if c.Components.EmailSenderService.MailType != "API" {
			return fmt.Errorf("email sender service mail_type must be API")
		}
		if c.Components.EmailSenderService.SenderDomain == "" {
			return fmt.Errorf("email sender service sender_domain is required when email sender is enabled")
		}
		if c.Components.EmailSenderService.MailAPIURL == "" {
			return fmt.Errorf("email sender service mail_api_url is required when email sender is enabled")
		}
		if c.Components.EmailSenderService.MailAPIKey == "" {
			return fmt.Errorf("email sender service mail_api_key is required when email sender is enabled")
		}
		if c.Images.EmailSenderServiceImage == "" {
			return fmt.Errorf("email sender service image is required when email sender is enabled")
		}
	}

	// Monitoring validation
	if c.Monitoring.Enabled {
		if c.Monitoring.URL == "" {
			return fmt.Errorf("monitoring url is required when monitoring is enabled")
		}
		if c.Monitoring.User == "" {
			return fmt.Errorf("monitoring user is required when monitoring is enabled")
		}
		if c.Monitoring.Password == "" {
			return fmt.Errorf("monitoring password is required when monitoring is enabled")
		}
		if c.Monitoring.SecurityKey == "" {
			return fmt.Errorf("monitoring security key is required when monitoring is enabled")
		}
	}

	// Image pull secrets validation
	if c.ImagePullSecrets.Enabled {
		if c.ImagePullSecrets.RegistryServer == "" {
			return fmt.Errorf("image pull secrets registry_server is required when image pull secrets is enabled")
		}
		if c.ImagePullSecrets.RegistryUsername == "" {
			return fmt.Errorf("image pull secrets registry_username is required when image pull secrets is enabled")
		}
		if c.ImagePullSecrets.RegistryPassword == "" {
			return fmt.Errorf("image pull secrets registry_password is required when image pull secrets is enabled")
		}
	}

	return nil
}

// AsValues converts Config into Helm chart values.
func (c *Config) AsValues() (*chartutil.Values, error) {
	configYAML, err := yaml.Marshal(c)
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
