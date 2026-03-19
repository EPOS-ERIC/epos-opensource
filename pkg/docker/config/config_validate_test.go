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
