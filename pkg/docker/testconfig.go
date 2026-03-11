package docker

import (
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

func newTestConfig(t *testing.T, name string) *config.EnvConfig {
	t.Helper()
	return &config.EnvConfig{
		Name:     name,
		Domain:   "localhost",
		Protocol: "http",
		Components: config.Components{
			PlatformGUI: config.PlatformGUI{BaseURL: "/", Port: 32000},
			Gateway:     config.Gateway{BaseURL: "/api/v1", Port: 33000},
			Backoffice: config.Backoffice{
				Enabled: false,
				GUI:     config.GUI{BaseURL: "/backoffice", Port: 34000},
				Service: config.Service{Auth: config.Auth{Enabled: false, OnlyAdmin: false}},
			},
			Converter: config.Converter{
				Enabled: false,
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
				Enabled: true,
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
				Enabled:         false,
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
