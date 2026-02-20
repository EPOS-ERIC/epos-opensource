package docker

import (
	"strings"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

func TestRenderOpts_Validate(t *testing.T) {
	tests := []struct {
		name         string
		opts         RenderOpts
		wantErr      bool
		errContains  string
		wantName     string
		wantConfig   bool
		checkNameSet bool
	}{
		{
			name:        "defaults config when nil and still requires name",
			opts:        RenderOpts{},
			wantErr:     true,
			errContains: "environment name is required",
			wantConfig:  true,
		},
		{
			name:       "uses name from config when name is empty",
			opts:       RenderOpts{Config: &config.EnvConfig{Name: "from-config"}},
			wantErr:    false,
			wantName:   "from-config",
			wantConfig: true,
		},
		{
			name:       "name argument overrides config name",
			opts:       RenderOpts{Name: "from-arg", Config: &config.EnvConfig{Name: "from-config"}},
			wantErr:    false,
			wantName:   "from-arg",
			wantConfig: true,
		},
		{
			name:        "returns error when resolved name is empty",
			opts:        RenderOpts{Config: &config.EnvConfig{}},
			wantErr:     true,
			errContains: "environment name is required",
		},
		{
			name:        "returns error for invalid name",
			opts:        RenderOpts{Name: "invalid/name", Config: &config.EnvConfig{}},
			wantErr:     true,
			errContains: "invalid name for environment",
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

			if tt.wantConfig && tt.opts.Config == nil {
				t.Fatal("Validate() did not initialize config")
			}

			if tt.checkNameSet && tt.opts.Name == "" {
				t.Fatal("Validate() did not resolve a name")
			}

			if tt.wantName != "" {
				if tt.opts.Name != tt.wantName {
					t.Fatalf("Validate() opts.Name = %q, want %q", tt.opts.Name, tt.wantName)
				}
				if tt.opts.Config == nil || tt.opts.Config.Name != tt.wantName {
					t.Fatalf("Validate() opts.Config.Name = %q, want %q", tt.opts.Config.Name, tt.wantName)
				}
			}
		})
	}
}
