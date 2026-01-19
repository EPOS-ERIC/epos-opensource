package config

import "github.com/EPOS-ERIC/epos-opensource/common"

type EnvConfig struct {
	Name       string        `yaml:"name"`
	Domain     string        `yaml:"domain"`
	Protocol   string        `yaml:"protocol"`
	Components Components    `yaml:"components"`
	Monitoring Monitoring    `yaml:"monitoring"`
	Images     common.Images `yaml:"images"`
}

type PlatformGUI struct {
	BaseURL string `yaml:"base_url"`
	Port    int    `yaml:"port"`
}

type Aai struct {
	Enabled         bool   `yaml:"enabled"`
	ServiceEndpoint string `yaml:"service_endpoint"`
	SecurityKey     string `yaml:"security_key"`
}

type SwaggerPage struct {
	Tile         string `yaml:"tile"`
	Version      string `yaml:"version"`
	ContactEmail string `yaml:"contact_email"`
}

type Gateway struct {
	Aai         Aai         `yaml:"aai"`
	BaseURL     string      `yaml:"base_url"`
	SwaggerPage SwaggerPage `yaml:"swagger_page"`
	Port        int         `yaml:"port"`
}

type GUI struct {
	BaseURL string `yaml:"base_url"`
	Port    int    `yaml:"port"`
}

type Auth struct {
	Enabled   bool `yaml:"enabled"`
	OnlyAdmin bool `yaml:"only_admin"`
}

type Service struct {
	Auth Auth `yaml:"auth"`
}

type Backoffice struct {
	Enabled bool    `yaml:"enabled"`
	GUI     GUI     `yaml:"gui"`
	Service Service `yaml:"service"`
}

type Converter struct {
	Enabled bool `yaml:"enabled"`
	Auth    Auth `yaml:"auth"`
}

type ResourcesService struct {
	Auth     Auth `yaml:"auth"`
	CacheTTL int  `yaml:"cache_ttl"`
}

type IngestorService struct {
	Auth Auth   `yaml:"auth"`
	Hash string `yaml:"hash"`
}

type ExternalAccessService struct {
	Auth Auth `yaml:"auth"`
}

type SharingService struct {
	Enabled bool `yaml:"enabled"`
	Auth    Auth `yaml:"auth"`
}

type Rabbitmq struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Vhost    string `yaml:"vhost"`
}

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

type Monitoring struct {
	Enabled  bool   `yaml:"enabled"`
	URL      string `yaml:"url"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}
