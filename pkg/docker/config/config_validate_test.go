package config_test

import (
	"strings"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

func TestEnvConfigValidate_RequiresCoreImages(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(cfg *config.EnvConfig)
		errContains string
	}{
		{
			name: "rabbitmq image is required",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Images.RabbitmqImage = ""
			},
			errContains: "rabbitmq image is required",
		},
		{
			name: "dataportal image is required",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Images.DataportalImage = ""
			},
			errContains: "dataportal image is required",
		},
		{
			name: "gateway image is required",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Images.GatewayImage = ""
			},
			errContains: "gateway image is required",
		},
		{
			name: "metadata database image is required",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Images.MetadataDatabaseImage = ""
			},
			errContains: "metadata database image is required",
		},
		{
			name: "resources service image is required",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Images.ResourcesServiceImage = ""
			},
			errContains: "resources service image is required",
		},
		{
			name: "ingestor service image is required",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Images.IngestorServiceImage = ""
			},
			errContains: "ingestor service image is required",
		},
		{
			name: "external access image is required",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Images.ExternalAccessImage = ""
			},
			errContains: "external access image is required",
		},
		{
			name: "metadata database published port must be within range when set",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.MetadataDatabase.PublishedPort = 70000
			},
			errContains: "metadata database published_port must be between 1 and 65535 when set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTestConfig(t, "test-env").Build()
			tt.mutate(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatalf("Validate() error = nil, want error containing %q", tt.errContains)
			}

			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("Validate() error = %q, want substring %q", err.Error(), tt.errContains)
			}
		})
	}
}

func TestEnvConfigValidate_DefaultConfigIsValidWithName(t *testing.T) {
	cfg := config.GetDefaultConfig()
	cfg.Name = "default-env"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestEnvConfigValidate_MetadataDatabasePublishedPortIsOptional(t *testing.T) {
	cfg := NewTestConfig(t, "test-env").Build()
	cfg.Components.MetadataDatabase.PublishedPort = 35432

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestEnvConfigValidate_AAI(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(cfg *config.EnvConfig)
		errContains string
	}{
		{
			name: "aai service requires gateway aai",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.AAIService.Enabled = true
			},
			errContains: "aai service is enabled but aai in the gateway is disabled",
		},
		{
			name: "aai service port is required when enabled",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.AAIService.Enabled = true
				cfg.Components.AAIService.Port = 0
			},
			errContains: "aai service port is required when aai service is enabled",
		},
		{
			name: "aai service port must be within range",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.AAIService.Enabled = true
				cfg.Components.AAIService.Port = 70000
			},
			errContains: "aai service port must be between 1 and 65535",
		},
		{
			name: "aai service name is required when enabled",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.AAIService.Enabled = true
				cfg.Components.AAIService.Name = ""
			},
			errContains: "aai service name is required when aai service is enabled",
		},
		{
			name: "aai service surname is required when enabled",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.AAIService.Enabled = true
				cfg.Components.AAIService.Surname = ""
			},
			errContains: "aai service surname is required when aai service is enabled",
		},
		{
			name: "aai service email is required when enabled",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.AAIService.Enabled = true
				cfg.Components.AAIService.Email = ""
			},
			errContains: "aai service email is required when aai service is enabled",
		},
		{
			name: "aai service password is required when enabled",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.AAIService.Enabled = true
				cfg.Components.AAIService.Password = ""
			},
			errContains: "aai service password is required when aai service is enabled",
		},
		{
			name: "aai service image is required when enabled",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.AAIService.Enabled = true
				cfg.Images.AAIServiceImage = ""
			},
			errContains: "aai service image is required when aai service is enabled",
		},
		{
			name: "gateway aai requires service endpoint",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.Gateway.AAI.ServiceEndpoint = ""
			},
			errContains: "aai service endpoint is required when aai is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTestConfig(t, "test-env").Build()
			tt.mutate(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatalf("Validate() error = nil, want error containing %q", tt.errContains)
			}

			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("Validate() error = %q, want substring %q", err.Error(), tt.errContains)
			}
		})
	}
}

func TestEnvConfigValidate_GatewayAAIWithExternalEndpointIsValid(t *testing.T) {
	cfg := NewTestConfig(t, "test-env").Build()
	cfg.Components.Gateway.AAI.Enabled = true
	cfg.Components.Gateway.AAI.ServiceEndpoint = "https://auth.example.com/oauth2/userinfo"
	cfg.Components.AAIService.Enabled = false

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestEnvConfigValidate_BackofficeRequiresAuth(t *testing.T) {
	cfg := NewTestConfig(t, "test-env").WithBackoffice(true).Build()
	cfg.Components.Backoffice.Service.Auth.Enabled = false

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "backoffice service auth must be enabled when backoffice is enabled") {
		t.Fatalf("Validate() error = %q, want substring %q", err.Error(), "backoffice service auth must be enabled when backoffice is enabled")
	}
}

func TestEnvConfigValidate_ServiceAuthRequiresGatewayAAI(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(cfg *config.EnvConfig)
		errContains string
	}{
		{
			name: "backoffice service auth requires gateway aai",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.Backoffice.Service.Auth.Enabled = true
			},
			errContains: "backoffice service auth requires gateway aai to be enabled",
		},
		{
			name: "converter service auth requires gateway aai",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.Converter.Auth.Enabled = true
			},
			errContains: "converter service auth requires gateway aai to be enabled",
		},
		{
			name: "resources service auth requires gateway aai",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.ResourcesService.Auth.Enabled = true
			},
			errContains: "resources service auth requires gateway aai to be enabled",
		},
		{
			name: "ingestor service auth requires gateway aai",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.IngestorService.Auth.Enabled = true
			},
			errContains: "ingestor service auth requires gateway aai to be enabled",
		},
		{
			name: "external access service auth requires gateway aai",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.ExternalAccessService.Auth.Enabled = true
			},
			errContains: "external access service auth requires gateway aai to be enabled",
		},
		{
			name: "sharing service auth requires gateway aai",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.SharingService.Auth.Enabled = true
			},
			errContains: "sharing service auth requires gateway aai to be enabled",
		},
		{
			name: "email sender service auth requires gateway aai",
			mutate: func(cfg *config.EnvConfig) {
				cfg.Components.EmailSenderService.Auth.Enabled = true
			},
			errContains: "email sender service auth requires gateway aai to be enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTestConfig(t, "test-env").Build()
			tt.mutate(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatalf("Validate() error = nil, want error containing %q", tt.errContains)
			}

			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("Validate() error = %q, want substring %q", err.Error(), tt.errContains)
			}
		})
	}
}

func TestEnvConfigValidate_ServiceAuthIsValidWithGatewayAAI(t *testing.T) {
	cfg := NewTestConfig(t, "test-env").Build()
	cfg.Components.Gateway.AAI.Enabled = true
	cfg.Components.Gateway.AAI.ServiceEndpoint = "https://auth.example.com/oauth2/userinfo"
	cfg.Components.ResourcesService.Auth.Enabled = true

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}
