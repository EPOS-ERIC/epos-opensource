package config

import (
	"strings"
	"testing"
)

func TestConfigValidate_ImageRequirements(t *testing.T) {
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
