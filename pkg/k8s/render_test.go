package k8s

import (
	"sort"
	"strings"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
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
			opts:       RenderOpts{Config: &config.Config{Name: "from-config"}},
			wantErr:    false,
			wantName:   "from-config",
			wantConfig: true,
		},
		{
			name:       "name argument overrides config name",
			opts:       RenderOpts{Name: "from-arg", Config: &config.Config{Name: "from-config"}},
			wantErr:    false,
			wantName:   "from-arg",
			wantConfig: true,
		},
		{
			name:        "returns error when resolved name is empty",
			opts:        RenderOpts{Config: &config.Config{}},
			wantErr:     true,
			errContains: "environment name is required",
		},
		{
			name:        "returns error for invalid name",
			opts:        RenderOpts{Name: "invalid/name", Config: &config.Config{}},
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

func TestRenderOpts_ConfigRender_OptionalComponents(t *testing.T) {
	tests := []struct {
		name          string
		envName       string
		mutate        func(cfg *config.Config)
		wantGateway   []string
		wantTemplates map[string]string
	}{
		{
			name:    "email sender enabled renders email sender chart and gateway flag",
			envName: "test-email-sender",
			mutate: func(cfg *config.Config) {
				cfg.Components.EmailSenderService.Enabled = true
				cfg.Components.EmailSenderService.Auth = config.Auth{
					Enabled:   true,
					OnlyAdmin: true,
				}
			},
			wantGateway: []string{`LOAD_EMAIL_SENDER_API: "true:true"`},
			wantTemplates: map[string]string{
				"templates/email-sender-service.yaml": "name: email-sender-service",
			},
		},
		{
			name:    "sharing enabled renders sharing chart and gateway flag",
			envName: "test-sharing",
			mutate: func(cfg *config.Config) {
				cfg.Components.SharingService.Enabled = true
			},
			wantGateway: []string{`LOAD_SHARING_API: "false:false"`},
			wantTemplates: map[string]string{
				"templates/sharing-service.yaml": "name: sharing-service",
			},
		},
		{
			name:    "backoffice enabled renders backoffice chart and gateway flag",
			envName: "test-backoffice",
			mutate: func(cfg *config.Config) {
				cfg.Components.Backoffice.Enabled = true
			},
			wantGateway: []string{`LOAD_BACKOFFICE_API: "false:false"`},
			wantTemplates: map[string]string{
				"templates/backoffice-service.yaml": "name: backoffice-service",
			},
		},
		{
			name:    "converter enabled renders converter charts and gateway flag",
			envName: "test-converter",
			mutate: func(cfg *config.Config) {
				cfg.Components.Converter.Enabled = true
			},
			wantGateway: []string{`LOAD_CONVERTER_API: "false:false"`},
			wantTemplates: map[string]string{
				"templates/converter-service.yaml": "name: converter-service",
				"templates/converter-routine.yaml": "name: converter-routine",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.GetDefaultConfig()
			cfg.Name = tt.envName
			tt.mutate(cfg)

			opts := RenderOpts{Config: cfg}
			if err := opts.Validate(); err != nil {
				t.Fatalf("Validate() returned error: %v", err)
			}

			files, err := opts.Config.Render()
			if err != nil {
				t.Fatalf("Config.Render() returned error: %v", err)
			}

			gatewayManifest := mustGetRenderedFileBySuffix(t, files, "templates/gateway.yaml")
			for _, want := range tt.wantGateway {
				if !strings.Contains(gatewayManifest, want) {
					t.Fatalf("gateway manifest missing expected content %q", want)
				}
			}

			for suffix, want := range tt.wantTemplates {
				manifest := mustGetRenderedFileBySuffix(t, files, suffix)
				if !strings.Contains(manifest, want) {
					t.Fatalf("rendered file %q missing expected content %q", suffix, want)
				}
			}
		})
	}
}

func mustGetRenderedFileBySuffix(t *testing.T, files map[string]string, suffix string) string {
	t.Helper()

	for path, content := range files {
		if strings.HasSuffix(path, suffix) {
			return content
		}
	}

	keys := make([]string, 0, len(files))
	for path := range files {
		keys = append(keys, path)
	}
	sort.Strings(keys)

	t.Fatalf("rendered file with suffix %q not found; got files: %v", suffix, keys)

	return ""
}
