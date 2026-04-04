package config

import (
	"strings"
	"testing"
)

const externalAAIUserinfoEndpoint = "https://auth.example.com/oauth2/userinfo"

func TestConfigValidate_Requirements(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(cfg *Config)
		wantErr     bool
		errContains string
	}{
		{
			name: "dataportal image is required",
			mutate: func(cfg *Config) {
				cfg.Images.DataportalImage = ""
			},
			wantErr:     true,
			errContains: "dataportal image is required",
		},
		{
			name: "gateway image is required",
			mutate: func(cfg *Config) {
				cfg.Images.GatewayImage = ""
			},
			wantErr:     true,
			errContains: "gateway image is required",
		},
		{
			name: "resources service image is required",
			mutate: func(cfg *Config) {
				cfg.Images.ResourcesServiceImage = ""
			},
			wantErr:     true,
			errContains: "resources service image is required",
		},
		{
			name: "ingestor service image is required",
			mutate: func(cfg *Config) {
				cfg.Images.IngestorServiceImage = ""
			},
			wantErr:     true,
			errContains: "ingestor service image is required",
		},
		{
			name: "external access image is required",
			mutate: func(cfg *Config) {
				cfg.Images.ExternalAccessImage = ""
			},
			wantErr:     true,
			errContains: "external access image is required",
		},
		{
			name: "rabbitmq image is required when rabbitmq is enabled",
			mutate: func(cfg *Config) {
				cfg.Components.Rabbitmq.Enabled = true
				cfg.Images.RabbitmqImage = ""
			},
			wantErr:     true,
			errContains: "rabbitmq image is required when rabbitmq is enabled",
		},
		{
			name: "metadata database image is required when metadata database is enabled",
			mutate: func(cfg *Config) {
				cfg.Components.MetadataDatabase.Enabled = true
				cfg.Images.MetadataDatabaseImage = ""
			},
			wantErr:     true,
			errContains: "metadata database image is required when metadata database is enabled",
		},
		{
			name: "rabbitmq image not required when rabbitmq is disabled",
			mutate: func(cfg *Config) {
				cfg.Components.Rabbitmq.Enabled = false
				cfg.Images.RabbitmqImage = ""
			},
			wantErr: false,
		},
		{
			name: "metadata database image not required when metadata database is disabled",
			mutate: func(cfg *Config) {
				cfg.Components.MetadataDatabase.Enabled = false
				cfg.Images.MetadataDatabaseImage = ""
			},
			wantErr: false,
		},
		{
			name: "aai service endpoint is required when aai is enabled",
			mutate: func(cfg *Config) {
				enableAAI(cfg)
				cfg.Components.Gateway.AAI.ServiceEndpoint = ""
			},
			wantErr:     true,
			errContains: "aai service endpoint is required when aai is enabled",
		},
		{
			name: "monitoring url is required when monitoring is enabled",
			mutate: func(cfg *Config) {
				enableMonitoring(cfg)
				cfg.Monitoring.URL = ""
			},
			wantErr:     true,
			errContains: "monitoring url is required when monitoring is enabled",
		},
		{
			name: "monitoring user is required when monitoring is enabled",
			mutate: func(cfg *Config) {
				enableMonitoring(cfg)
				cfg.Monitoring.User = ""
			},
			wantErr:     true,
			errContains: "monitoring user is required when monitoring is enabled",
		},
		{
			name: "monitoring password is required when monitoring is enabled",
			mutate: func(cfg *Config) {
				enableMonitoring(cfg)
				cfg.Monitoring.Password = ""
			},
			wantErr:     true,
			errContains: "monitoring password is required when monitoring is enabled",
		},
		{
			name: "monitoring security key is required when monitoring is enabled",
			mutate: func(cfg *Config) {
				enableMonitoring(cfg)
				cfg.Monitoring.SecurityKey = ""
			},
			wantErr:     true,
			errContains: "monitoring security key is required when monitoring is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetDefaultConfig()
			tt.mutate(cfg)

			err := cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Validate() error = nil, want error containing %q", tt.errContains)
				}

				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("Validate() error = %q, want substring %q", err.Error(), tt.errContains)
				}

				return
			}

			if err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestConfigValidate_DefaultConfigIsValid(t *testing.T) {
	cfg := GetDefaultConfig()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestConfigValidate_AAI(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(cfg *Config)
		errContains string
	}{
		{
			name: "aai service requires gateway aai",
			mutate: func(cfg *Config) {
				cfg.Components.AAIService.Enabled = true
			},
			errContains: "aai service is enabled but aai in the gateway is disabled",
		},
		{
			name: "aai service name is required when enabled",
			mutate: func(cfg *Config) {
				enableAAI(cfg)
				cfg.Components.AAIService.Enabled = true
				cfg.Components.AAIService.Name = ""
			},
			errContains: "aai service name is required when aai service is enabled",
		},
		{
			name: "aai service surname is required when enabled",
			mutate: func(cfg *Config) {
				enableAAI(cfg)
				cfg.Components.AAIService.Enabled = true
				cfg.Components.AAIService.Surname = ""
			},
			errContains: "aai service surname is required when aai service is enabled",
		},
		{
			name: "aai service email is required when enabled",
			mutate: func(cfg *Config) {
				enableAAI(cfg)
				cfg.Components.AAIService.Enabled = true
				cfg.Components.AAIService.Email = ""
			},
			errContains: "aai service email is required when aai service is enabled",
		},
		{
			name: "aai service password is required when enabled",
			mutate: func(cfg *Config) {
				enableAAI(cfg)
				cfg.Components.AAIService.Enabled = true
				cfg.Components.AAIService.Password = ""
			},
			errContains: "aai service password is required when aai service is enabled",
		},
		{
			name: "aai service image is required when enabled",
			mutate: func(cfg *Config) {
				enableAAI(cfg)
				cfg.Components.AAIService.Enabled = true
				cfg.Images.AAIServiceImage = ""
			},
			errContains: "aai service image is required when aai service is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetDefaultConfig()
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

func TestConfigValidate_GatewayAAIWithExternalEndpointIsValid(t *testing.T) {
	cfg := GetDefaultConfig()
	cfg.Components.Gateway.AAI.Enabled = true
	cfg.Components.Gateway.AAI.ServiceEndpoint = externalAAIUserinfoEndpoint
	cfg.Components.AAIService.Enabled = false

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestConfigValidate_BackofficeRequiresAuth(t *testing.T) {
	cfg := GetDefaultConfig()
	cfg.Components.Backoffice.Enabled = true
	cfg.Components.Backoffice.Service.Auth.Enabled = false

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "backoffice service auth must be enabled when backoffice is enabled") {
		t.Fatalf("Validate() error = %q, want substring %q", err.Error(), "backoffice service auth must be enabled when backoffice is enabled")
	}
}

func TestConfigValidate_ServiceAuthRequiresGatewayAAI(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(cfg *Config)
		errContains string
	}{
		{
			name: "backoffice service auth requires gateway aai",
			mutate: func(cfg *Config) {
				cfg.Components.Backoffice.Service.Auth.Enabled = true
			},
			errContains: "backoffice service auth requires gateway aai to be enabled",
		},
		{
			name: "converter service auth requires gateway aai",
			mutate: func(cfg *Config) {
				cfg.Components.Converter.Auth.Enabled = true
			},
			errContains: "converter service auth requires gateway aai to be enabled",
		},
		{
			name: "resources service auth requires gateway aai",
			mutate: func(cfg *Config) {
				cfg.Components.ResourcesService.Auth.Enabled = true
			},
			errContains: "resources service auth requires gateway aai to be enabled",
		},
		{
			name: "ingestor service auth requires gateway aai",
			mutate: func(cfg *Config) {
				cfg.Components.IngestorService.Auth.Enabled = true
			},
			errContains: "ingestor service auth requires gateway aai to be enabled",
		},
		{
			name: "external access service auth requires gateway aai",
			mutate: func(cfg *Config) {
				cfg.Components.ExternalAccessService.Auth.Enabled = true
			},
			errContains: "external access service auth requires gateway aai to be enabled",
		},
		{
			name: "sharing service auth requires gateway aai",
			mutate: func(cfg *Config) {
				cfg.Components.SharingService.Auth.Enabled = true
			},
			errContains: "sharing service auth requires gateway aai to be enabled",
		},
		{
			name: "email sender service auth requires gateway aai",
			mutate: func(cfg *Config) {
				cfg.Components.EmailSenderService.Auth.Enabled = true
			},
			errContains: "email sender service auth requires gateway aai to be enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetDefaultConfig()
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

func TestConfigValidate_ServiceAuthIsValidWithGatewayAAI(t *testing.T) {
	cfg := GetDefaultConfig()
	cfg.Components.Gateway.AAI.Enabled = true
	cfg.Components.Gateway.AAI.ServiceEndpoint = externalAAIUserinfoEndpoint
	cfg.Components.ResourcesService.Auth.Enabled = true

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func enableAAI(cfg *Config) {
	cfg.Components.Gateway.AAI.Enabled = true
	cfg.Components.Gateway.AAI.ServiceEndpoint = "https://aai.example.com"
}

func enableMonitoring(cfg *Config) {
	cfg.Monitoring.Enabled = true
	cfg.Monitoring.URL = "https://monitoring.example.com"
	cfg.Monitoring.User = "monitor-user"
	cfg.Monitoring.Password = "monitor-password"
	cfg.Monitoring.SecurityKey = "test-security-key"
}
