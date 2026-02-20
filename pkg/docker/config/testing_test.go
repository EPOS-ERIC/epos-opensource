package config_test

import (
	"strings"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

type TestConfigBuilder struct {
	t                 *testing.T
	name              string
	domain            string
	protocol          string
	guiPort           int
	gatewayPort       int
	backofficeEnabled bool
	converterEnabled  bool
	sharingEnabled    bool
	emailEnabled      bool
}

func NewTestConfig(t *testing.T, name string) *TestConfigBuilder {
	return &TestConfigBuilder{
		t:                 t,
		name:              name,
		domain:            "localhost",
		protocol:          "http",
		guiPort:           32000,
		gatewayPort:       33000,
		backofficeEnabled: false,
		converterEnabled:  false,
		sharingEnabled:    true,
		emailEnabled:      false,
	}
}

func (b *TestConfigBuilder) WithPorts(guiPort, gatewayPort int) *TestConfigBuilder {
	b.guiPort = guiPort
	b.gatewayPort = gatewayPort
	return b
}

func (b *TestConfigBuilder) WithBackoffice(enabled bool) *TestConfigBuilder {
	b.backofficeEnabled = enabled
	return b
}

func (b *TestConfigBuilder) WithConverter(enabled bool) *TestConfigBuilder {
	b.converterEnabled = enabled
	return b
}

func (b *TestConfigBuilder) WithSharing(enabled bool) *TestConfigBuilder {
	b.sharingEnabled = enabled
	return b
}

func (b *TestConfigBuilder) WithEmailSender(enabled bool) *TestConfigBuilder {
	b.emailEnabled = enabled
	return b
}

func (b *TestConfigBuilder) Build() *config.EnvConfig {
	return &config.EnvConfig{
		Name:     b.name,
		Domain:   b.domain,
		Protocol: b.protocol,
		Components: config.Components{
			PlatformGUI: config.PlatformGUI{
				BaseURL: "/",
				Port:    b.guiPort,
			},
			Gateway: config.Gateway{
				BaseURL: "/api/v1",
				Port:    b.gatewayPort,
			},
			Backoffice: config.Backoffice{
				Enabled: b.backofficeEnabled,
				GUI:     config.GUI{BaseURL: "/backoffice", Port: 34000},
				Service: config.Service{Auth: config.Auth{Enabled: false, OnlyAdmin: false}},
			},
			Converter: config.Converter{
				Enabled: b.converterEnabled,
				Auth:    config.Auth{Enabled: false, OnlyAdmin: false},
			},
			ResourcesService: config.ResourcesService{
				Auth:     config.Auth{Enabled: false, OnlyAdmin: false},
				CacheTTL: 5000,
			},
			IngestorService: config.IngestorService{
				Auth: config.Auth{Enabled: false, OnlyAdmin: false},
				Hash: "FA9BEB99E4029AD5A6615399E7BBAE21356086B3",
			},
			ExternalAccessService: config.ExternalAccessService{
				Auth: config.Auth{Enabled: false, OnlyAdmin: false},
			},
			SharingService: config.SharingService{
				Enabled: b.sharingEnabled,
				Auth:    config.Auth{Enabled: false, OnlyAdmin: false},
			},
			Rabbitmq: config.Rabbitmq{
				Host:     "rabbitmq",
				Username: "rabbitmq-user",
				Password: "changeme",
				Vhost:    "changeme",
			},
			MetadataDatabase: config.MetadataDatabase{
				User:                   "metadatauser",
				Password:               "changeme",
				Host:                   "metadata-database",
				Port:                   5432,
				DBName:                 "cerif",
				ConnectionPoolInitSize: 5,
				ConnectionPoolMinSize:  5,
				ConnectionPoolMaxSize:  15,
			},
			EmailSenderService: config.EmailSenderService{
				Enabled:         b.emailEnabled,
				EnvironmentType: "development",
				Sender:          "data-portal",
				SenderName:      "EPOS Platform Opensource",
				MailType:        "API",
				SenderDomain:    "your.domain.com",
				MailHost:        "mail.domain.com",
				MailUser:        "user@email.com",
				MailPassword:    "changeme",
				DevEmails:       "foo.bar@somewhere.com",
				MailAPIURL:      "https://api.example.email/",
				MailAPIKey:      "changeme",
			},
		},
		Monitoring: config.Monitoring{
			Enabled: false,
		},
		Images: common.Images{
			RabbitmqImage:           "rabbitmq:3.13.7-management",
			DataportalImage:         "epos/data-portal:latest",
			GatewayImage:            "ghcr.io/epos-eric/epos-api-gateway:latest",
			MetadataDatabaseImage:   "ghcr.io/epos-eric/metadata-database/deploy:latest",
			ResourcesServiceImage:   "ghcr.io/epos-eric/resources-service:latest",
			IngestorServiceImage:    "ghcr.io/epos-eric/ingestor-service:latest",
			ExternalAccessImage:     "ghcr.io/epos-eric/external-access-service:latest",
			ConverterServiceImage:   "ghcr.io/epos-eric/converter-service-go:latest",
			ConverterRoutineImage:   "ghcr.io/epos-eric/converter-routine-go:latest",
			BackofficeServiceImage:  "ghcr.io/epos-eric/backoffice-service:latest",
			BackofficeUIImage:       "epos/backoffice-ui:latest",
			EmailSenderServiceImage: "ghcr.io/epos-eric/email-sender-service:latest",
			SharingServiceImage:     "ghcr.io/epos-eric/sharing-service:latest",
		},
	}
}

func MustRender(t *testing.T, cfg *config.EnvConfig) map[string]string {
	t.Helper()
	got, err := cfg.Render()
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}
	return got
}

func ContentContains(t *testing.T, content, fileName string, expected []string) {
	t.Helper()
	for _, s := range expected {
		if !strings.Contains(content, s) {
			t.Errorf("File %q missing expected content %q", fileName, s)
		}
	}
}

func ContentExcludes(t *testing.T, content, fileName string, unexpected []string) {
	t.Helper()
	for _, s := range unexpected {
		if strings.Contains(content, s) {
			t.Errorf("File %q should NOT contain %q", fileName, s)
		}
	}
}
