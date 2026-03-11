package docker

import (
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

func TestDeployOpts_Validate(t *testing.T) {
	validConfig := newTestConfig(t, "test-deploy")
	invalidConfig := &config.EnvConfig{}
	invalidNameConfig := newTestConfig(t, "787/(//&&$%$%&)(/((=??=)))")

	tests := []struct {
		name    string
		opts    DeployOpts
		wantErr bool
	}{
		{
			name:    "Nil config returns error",
			opts:    DeployOpts{Config: nil},
			wantErr: true,
		},
		{
			name:    "Invalid config returns error",
			opts:    DeployOpts{Config: invalidConfig},
			wantErr: true,
		},
		{
			name:    "Invalid name returns error",
			opts:    DeployOpts{Config: invalidNameConfig},
			wantErr: true,
		},
		{
			name:    "Valid opts",
			opts:    DeployOpts{Config: validConfig},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
