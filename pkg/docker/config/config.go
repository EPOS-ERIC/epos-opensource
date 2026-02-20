package config

import (
	_ "embed"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var defaultConfig []byte

// GetDefaultConfig returns a parsed copy of the embedded default Docker configuration.
func GetDefaultConfig() *EnvConfig {
	var config EnvConfig
	err := yaml.Unmarshal(defaultConfig, &config)
	if err != nil {
		panic(fmt.Errorf("error unmarshalling default config yaml: %w", err))
	}
	return &config
}

// GetDefaultConfigBytes returns the raw embedded default Docker configuration bytes.
func GetDefaultConfigBytes() []byte {
	return defaultConfig
}

// LoadConfig loads a Docker configuration from a YAML file.
func LoadConfig(path string) (*EnvConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config EnvConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config yaml: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return &config, nil
}

// Save writes the current configuration as YAML to path.
func (e *EnvConfig) Save(path string) error {
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

// CheckForUpdates checks configured container images for newer tags.
func (e *EnvConfig) CheckForUpdates() ([]display.ImageUpdateInfo, error) {
	imgs := map[string]string{
		"Rabbitmq":                e.Images.RabbitmqImage,
		"Platform UI":             e.Images.DataportalImage,
		"Gateway":                 e.Images.GatewayImage,
		"Metadata Database":       e.Images.MetadataDatabaseImage,
		"Resources Service":       e.Images.ResourcesServiceImage,
		"Ingestor Service":        e.Images.IngestorServiceImage,
		"External Access Service": e.Images.ExternalAccessImage,
	}

	if e.Components.Converter.Enabled {
		imgs["Converter Service"] = e.Images.ConverterServiceImage
		imgs["Converter Routine"] = e.Images.ConverterRoutineImage
	}

	if e.Components.Backoffice.Enabled {
		imgs["Backoffice Service"] = e.Images.BackofficeServiceImage
		imgs["Backoffice UI"] = e.Images.BackofficeUIImage
	}

	if e.Components.EmailSenderService.Enabled {
		imgs["Email Sender Service"] = e.Images.EmailSenderServiceImage
	}

	if e.Components.SharingService.Enabled {
		imgs["Sharing Service"] = e.Images.SharingServiceImage
	}

	updates, err := common.CheckEnvForUpdates(imgs)
	if err != nil {
		return nil, fmt.Errorf("error checking for updates: %w", err)
	}

	return updates, nil
}

// EnsurePortsFree verifies default service ports are available and auto-selects free ones when needed.
func (e *EnvConfig) EnsurePortsFree() error {
	if !e.UsingDefaultPorts() {
		return nil // User specified custom ports, use them as-is
	}

	// Check and auto-select ports for default configuration
	if free, err := common.IsPortFree(e.Components.PlatformGUI.Port); err != nil {
		display.Warn("error checking availability of port for platform GUI: %d. Continuing anyway", e.Components.PlatformGUI.Port)
	} else if !free {
		freePort, err := common.FindFreePort()
		if err != nil {
			return fmt.Errorf("error finding free port for platform GUI: %w", err)
		}
		display.Info("Port %d for platform GUI is already used, using new random free port: %d", e.Components.PlatformGUI.Port, freePort)
		e.Components.PlatformGUI.Port = freePort
	}

	if free, err := common.IsPortFree(e.Components.Gateway.Port); err != nil {
		display.Warn("error checking availability of port for gateway: %d. Continuing anyway", e.Components.Gateway.Port)
	} else if !free {
		freePort, err := common.FindFreePort()
		if err != nil {
			return fmt.Errorf("error finding free port for gateway: %w", err)
		}
		display.Info("Port %d for gateway is already used, using new random free port: %d", e.Components.Gateway.Port, freePort)
		e.Components.Gateway.Port = freePort
	}

	if e.Components.Backoffice.Enabled {
		if free, err := common.IsPortFree(e.Components.Backoffice.GUI.Port); err != nil {
			display.Warn("error checking availability of port for backoffice: %d. Continuing anyway", e.Components.Backoffice.GUI.Port)
		} else if !free {
			freePort, err := common.FindFreePort()
			if err != nil {
				return fmt.Errorf("error finding free port for backoffice: %w", err)
			}
			display.Info("Port %d for backoffice is already used, using new random free port: %d", e.Components.Backoffice.GUI.Port, freePort)
			e.Components.Backoffice.GUI.Port = freePort
		}
	}

	return nil
}

// Validate checks whether the configuration contains all required values.
func (e *EnvConfig) Validate() error {
	// Basic required fields
	if e.Name == "" {
		return fmt.Errorf("environment name is required")
	}
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
	if e.Components.PlatformGUI.Port == 0 {
		return fmt.Errorf("platform gui port is required")
	}
	if e.Components.PlatformGUI.Port < 1 || e.Components.PlatformGUI.Port > 65535 {
		return fmt.Errorf("platform gui port must be between 1 and 65535")
	}

	// Gateway validation
	if e.Components.Gateway.Port == 0 {
		return fmt.Errorf("gateway port is required")
	}
	if e.Components.Gateway.Port < 1 || e.Components.Gateway.Port > 65535 {
		return fmt.Errorf("gateway port must be between 1 and 65535")
	}
	if e.Components.Gateway.BaseURL == "" {
		return fmt.Errorf("gateway base url is required")
	}
	if !strings.HasPrefix(e.Components.Gateway.BaseURL, "/") {
		return fmt.Errorf("gateway base url must start with /")
	}
	if !strings.HasSuffix(e.Components.Gateway.BaseURL, "/api/v1") {
		return fmt.Errorf("gateway base url must end with /api/v1")
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
		if e.Components.Backoffice.GUI.Port == 0 {
			return fmt.Errorf("backoffice port is required when backoffice is enabled")
		}
		if e.Components.Backoffice.GUI.Port < 1 || e.Components.Backoffice.GUI.Port > 65535 {
			return fmt.Errorf("backoffice port must be between 1 and 65535")
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
