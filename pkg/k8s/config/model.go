package config

import "github.com/EPOS-ERIC/epos-opensource/common"

// Config represents the full Kubernetes environment values schema.
type Config struct {
	Name               string           `yaml:"name"`
	Domain             string           `yaml:"domain"`
	Protocol           string           `yaml:"protocol"`
	TLSEnabled         bool             `yaml:"tls_enabled"`
	URLPrefixNamespace bool             `yaml:"url_prefix_namespace"`
	Components         Components       `yaml:"components"`
	Jobs               Jobs             `yaml:"jobs"`
	Monitoring         Monitoring       `yaml:"monitoring"`
	ImagePullSecrets   ImagePullSecrets `yaml:"image_pull_secrets"`
	Images             common.Images    `yaml:"images"`
}

// PlatformGUI configures the platform GUI endpoint.
type PlatformGUI struct {
	BaseURL string `yaml:"base_url"`
}

// Aai configures AAI integration for gateway services.
type Aai struct {
	Enabled         bool   `yaml:"enabled"`
	ServiceEndpoint string `yaml:"service_endpoint"`
	SecurityKey     string `yaml:"security_key"`
}

// SwaggerPage configures API documentation metadata.
type SwaggerPage struct {
	Tile         string `yaml:"tile"`
	Version      string `yaml:"version"`
	ContactEmail string `yaml:"contact_email"`
}

// Gateway configures gateway URLs and auth integrations.
type Gateway struct {
	Aai         Aai         `yaml:"aai"`
	BaseURL     string      `yaml:"base_url"`
	SwaggerPage SwaggerPage `yaml:"swagger_page"`
}

// GUI configures a generic GUI service endpoint.
type GUI struct {
	BaseURL string `yaml:"base_url"`
}

// Auth configures auth switches for a service.
type Auth struct {
	Enabled   bool `yaml:"enabled"`
	OnlyAdmin bool `yaml:"only_admin"`
}

// Service defines common service settings.
type Service struct {
	Auth Auth `yaml:"auth"`
}

// Backoffice configures backoffice UI and service behavior.
type Backoffice struct {
	Enabled bool    `yaml:"enabled"`
	GUI     GUI     `yaml:"gui"`
	Service Service `yaml:"service"`
}

// Converter configures the converter service.
type Converter struct {
	Enabled bool `yaml:"enabled"`
	Auth    Auth `yaml:"auth"`
}

// ResourcesService configures the resources service.
type ResourcesService struct {
	Auth     Auth `yaml:"auth"`
	CacheTTL int  `yaml:"cache_ttl"`
}

// IngestorService configures the ingestor service.
type IngestorService struct {
	Auth Auth   `yaml:"auth"`
	Hash string `yaml:"hash"`
}

// ExternalAccessService configures the external access service.
type ExternalAccessService struct {
	Auth Auth `yaml:"auth"`
}

// SharingService configures the sharing service.
type SharingService struct {
	Enabled bool `yaml:"enabled"`
	Auth    Auth `yaml:"auth"`
}

// Rabbitmq configures RabbitMQ connection details.
type Rabbitmq struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Vhost    string `yaml:"vhost"`
}

// MetadataDatabase configures metadata database connection and pooling.
type MetadataDatabase struct {
	User                   string `yaml:"user"`
	Password               string `yaml:"password"`
	Host                   string `yaml:"host"`
	Port                   int    `yaml:"port"`
	DBName                 string `yaml:"db_name"`
	ConnectionPoolInitSize int    `yaml:"connection_pool_init_size"`
	ConnectionPoolMinSize  int    `yaml:"connection_pool_min_size"`
	ConnectionPoolMaxSize  int    `yaml:"connection_pool_max_size"`
}

// EmailSenderService configures email sender service settings.
type EmailSenderService struct {
	Enabled         bool   `yaml:"enabled"`
	EnvironmentType string `yaml:"environment_type"`
	Sender          string `yaml:"sender"`
	SenderName      string `yaml:"sender_name"`
	MailType        string `yaml:"mail_type"`
	SenderDomain    string `yaml:"sender_domain"`
	MailHost        string `yaml:"mail_host"`
	MailUser        string `yaml:"mail_user"`
	MailPassword    string `yaml:"mail_password"`
	DevEmails       string `yaml:"dev_emails"`
	MailAPIURL      string `yaml:"mail_api_url"`
	MailAPIKey      string `yaml:"mail_api_key"`
}

// InitDBJob configures the init-db Kubernetes job.
type InitDBJob struct {
	Enabled               bool   `yaml:"enabled"`
	Image                 string `yaml:"image"`
	BackoffLimit          int    `yaml:"backoff_limit"`
	ActiveDeadlineSeconds int    `yaml:"active_deadline_seconds"`
}

// PluginPopulatorJob configures the plugin-populator Kubernetes job.
type PluginPopulatorJob struct {
	Enabled               bool   `yaml:"enabled"`
	Image                 string `yaml:"image"`
	Tag                   string `yaml:"tag"`
	HookWeight            string `yaml:"hook_weight"`
	BackoffLimit          int    `yaml:"backoff_limit"`
	ActiveDeadlineSeconds int    `yaml:"active_deadline_seconds"`
}

// MetadataPopulatorJob configures the metadata-populator Kubernetes job.
type MetadataPopulatorJob struct {
	Enabled               bool   `yaml:"enabled"`
	Image                 string `yaml:"image"`
	BackoffLimit          int    `yaml:"backoff_limit"`
	ActiveDeadlineSeconds int    `yaml:"active_deadline_seconds"`
	MaxParallel           int    `yaml:"max_parallel"`
}

// Jobs groups all optional Kubernetes jobs settings.
type Jobs struct {
	Enabled           bool                 `yaml:"enabled"`
	InitDB            InitDBJob            `yaml:"init_db"`
	PluginPopulator   PluginPopulatorJob   `yaml:"plugin_populator"`
	MetadataPopulator MetadataPopulatorJob `yaml:"metadata_populator"`
}

// Components groups all service-level component settings.
type Components struct {
	PlatformGUI           PlatformGUI           `yaml:"platform_gui"`
	Gateway               Gateway               `yaml:"gateway"`
	Backoffice            Backoffice            `yaml:"backoffice"`
	Converter             Converter             `yaml:"converter"`
	ResourcesService      ResourcesService      `yaml:"resources_service"`
	IngestorService       IngestorService       `yaml:"ingestor_service"`
	ExternalAccessService ExternalAccessService `yaml:"external_access_service"`
	SharingService        SharingService        `yaml:"sharing_service"`
	Rabbitmq              Rabbitmq              `yaml:"rabbitmq"`
	MetadataDatabase      MetadataDatabase      `yaml:"metadata_database"`
	EmailSenderService    EmailSenderService    `yaml:"email_sender_service"`
}

// Monitoring configures optional monitoring integration.
type Monitoring struct {
	Enabled  bool   `yaml:"enabled"`
	URL      string `yaml:"url"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// ImagePullSecrets configures optional private registry pull secret generation.
type ImagePullSecrets struct {
	Enabled          bool   `yaml:"enabled"`
	RegistryServer   string `yaml:"registry_server"`
	RegistryUsername string `yaml:"registry_username"`
	RegistryPassword string `yaml:"registry_password"`
	RegistryEmail    string `yaml:"registry_email"`
}

// String returns a concise string representation of key Config fields.
func (c *Config) String() string {
	if c == nil {
		return ""
	}

	out := "name=" + c.Name + ", domain=" + c.Domain + ", protocol=" + c.Protocol

	if c.TLSEnabled {
		out += ", tls_enabled=true"
	} else {
		out += ", tls_enabled=false"
	}

	if c.URLPrefixNamespace {
		out += ", url_prefix_namespace=true"
	} else {
		out += ", url_prefix_namespace=false"
	}

	return out
}
