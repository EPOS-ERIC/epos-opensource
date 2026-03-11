package docker

import (
	"strings"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

func TestUpdateOpts_Validate(t *testing.T) {
	invalidConfig := &config.EnvConfig{}

	tests := []struct {
		name        string
		opts        UpdateOpts
		wantErr     bool
		errContains string
	}{
		{
			name:    "Empty environment name",
			opts:    UpdateOpts{OldEnvName: ""},
			wantErr: true,
		},
		{
			name:    "Environment does not exist",
			opts:    UpdateOpts{OldEnvName: "does-not-exist"},
			wantErr: true,
		},
		{
			name:    "Reset with custom config returns error",
			opts:    UpdateOpts{OldEnvName: "test", Reset: true, NewConfig: invalidConfig},
			wantErr: true,
		},
		{
			name:        "Custom config name must match target environment name",
			opts:        UpdateOpts{OldEnvName: "test", NewConfig: &config.EnvConfig{Name: "different-name"}},
			wantErr:     true,
			errContains: "must match environment name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.errContains)
			}
		})
	}
}
