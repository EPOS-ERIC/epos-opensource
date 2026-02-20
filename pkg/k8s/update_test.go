package k8s

import (
	"strings"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
)

func TestUpdateOpts_Validate(t *testing.T) {
	tests := []struct {
		name        string
		opts        UpdateOpts
		wantErr     bool
		errContains string
	}{
		{
			name:        "Empty environment name",
			opts:        UpdateOpts{OldEnvName: ""},
			wantErr:     true,
			errContains: "old environment name is required",
		},
		{
			name:        "Reset with custom config returns error",
			opts:        UpdateOpts{OldEnvName: "test", Reset: true, NewConfig: &config.Config{}},
			wantErr:     true,
			errContains: "reset and new config are mutually exclusive",
		},
		{
			name:        "Custom config name must match target environment name",
			opts:        UpdateOpts{OldEnvName: "test", NewConfig: &config.Config{Name: "different-name"}},
			wantErr:     true,
			errContains: "must match environment name",
		},
		{
			name:        "Force update still disallows renaming",
			opts:        UpdateOpts{OldEnvName: "test", Force: true, NewConfig: &config.Config{Name: "different-name"}},
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
