package config_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
)

const (
	externalAAIUserinfoEndpoint = "https://auth.example.com/oauth2/userinfo"
)

func TestConfigRender(t *testing.T) {
	tests := []struct {
		name         string
		mutate       func(cfg *config.Config)
		wantFiles    []string
		wantContains map[string][]string
		notContains  map[string][]string
	}{
		{
			name: "default config renders core manifests with optional components disabled",
			mutate: func(cfg *config.Config) {
				cfg.Name = "test-default"
			},
			wantFiles: []string{
				"templates/dataportal.yaml",
				"templates/gateway.yaml",
				"templates/rabbitmq.yaml",
				"templates/metadata-database.yaml",
				"templates/resources-service.yaml",
			},
			wantContains: map[string][]string{
				"templates/dataportal.yaml":        {"name: dataportal", `AUTH_ROOT_URL: "http://localhost/test-default/aai"`},
				"templates/gateway.yaml":           {"name: gateway", `IS_AAI_ENABLED: "false"`},
				"templates/rabbitmq.yaml":          {"name: rabbitmq"},
				"templates/metadata-database.yaml": {"name: metadata-database"},
				"templates/resources-service.yaml": {"name: resources-service", `MONITORING: "false"`},
			},
			notContains: map[string][]string{
				"templates/gateway.yaml": {
					"LOAD_BACKOFFICE_API:",
					"LOAD_CONVERTER_API:",
					"LOAD_EMAIL_SENDER_API:",
					"LOAD_SHARING_API:",
					"SECURITY_KEY:",
				},
				"templates/resources-service.yaml": {
					"MONITORING_URL:",
					"MONITORING_USER:",
					"MONITORING_PWD:",
				},
				"templates/aai-service.yaml":          {"name: aai-service"},
				"templates/backoffice-service.yaml":   {"name: backoffice-service"},
				"templates/backoffice-ui.yaml":        {"name: backoffice-ui"},
				"templates/converter-service.yaml":    {"name: converter-service"},
				"templates/converter-routine.yaml":    {"name: converter-routine"},
				"templates/email-sender-service.yaml": {"name: email-sender-service"},
				"templates/sharing-service.yaml":      {"name: sharing-service"},
			},
		},
		{
			name: "backoffice enabled renders backoffice manifests and gateway flag",
			mutate: func(cfg *config.Config) {
				cfg.Name = "test-backoffice"
				cfg.Components.Backoffice.Enabled = true
			},
			wantContains: map[string][]string{
				"templates/gateway.yaml":            {`LOAD_BACKOFFICE_API: "false:false"`},
				"templates/backoffice-service.yaml": {"name: backoffice-service"},
				"templates/backoffice-ui.yaml":      {"name: backoffice-ui", `AUTH_ROOT_URL: "http://localhost/test-backoffice/aai"`},
			},
		},
		{
			name: "converter enabled renders converter manifests and gateway flag",
			mutate: func(cfg *config.Config) {
				cfg.Name = "test-converter"
				cfg.Components.Converter.Enabled = true
			},
			wantContains: map[string][]string{
				"templates/gateway.yaml":           {`LOAD_CONVERTER_API: "false:false"`},
				"templates/converter-service.yaml": {"name: converter-service"},
				"templates/converter-routine.yaml": {"name: converter-routine"},
			},
		},
		{
			name: "sharing enabled renders sharing manifest and gateway flag",
			mutate: func(cfg *config.Config) {
				cfg.Name = "test-sharing"
				cfg.Components.SharingService.Enabled = true
			},
			wantContains: map[string][]string{
				"templates/gateway.yaml":         {`LOAD_SHARING_API: "false:false"`},
				"templates/sharing-service.yaml": {"name: sharing-service"},
			},
		},
		{
			name: "email sender enabled renders manifest and gateway auth flag",
			mutate: func(cfg *config.Config) {
				cfg.Name = "test-email-auth"
				cfg.Components.EmailSenderService.Enabled = true
				cfg.Components.EmailSenderService.Auth = config.Auth{
					Enabled:   true,
					OnlyAdmin: true,
				}
			},
			wantContains: map[string][]string{
				"templates/gateway.yaml":              {`LOAD_EMAIL_SENDER_API: "true:true"`},
				"templates/email-sender-service.yaml": {"name: email-sender-service"},
			},
		},
		{
			name: "embedded aai enabled renders manifest and gateway settings",
			mutate: func(cfg *config.Config) {
				cfg.Name = "test-aai"
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.AAIService.Enabled = true
			},
			wantContains: map[string][]string{
				"templates/gateway.yaml": {
					`IS_AAI_ENABLED: "true"`,
					`AAI_SERVICE_ENDPOINT: "http://aai-service:8080/oauth2/userinfo"`,
					"wait-for-aai-service",
				},
				"templates/aai-service.yaml": {
					"name: aai-service",
					"kind: Ingress",
					"path: /test-aai/aai(/|$)(.*)",
					`INITIAL_ADMIN_NAME: "EPOS"`,
					`INITIAL_ADMIN_SURNAME: "User"`,
					`INITIAL_ADMIN_EMAIL: "epos@epos.eu"`,
					`INITIAL_ADMIN_PASSWORD: "epos"`,
					`APP_CORS_ALLOW_ORIGIN: "*"`,
				},
				"templates/pvc.yaml": {"name: aai"},
			},
		},
		{
			name: "embedded aai disabled supports external endpoint without manifest",
			mutate: func(cfg *config.Config) {
				cfg.Name = "test-external-aai"
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.Gateway.AAI.ServiceEndpoint = externalAAIUserinfoEndpoint
			},
			wantContains: map[string][]string{
				"templates/gateway.yaml": {
					`IS_AAI_ENABLED: "true"`,
					`AAI_SERVICE_ENDPOINT: "` + externalAAIUserinfoEndpoint + `"`,
				},
			},
			notContains: map[string][]string{
				"templates/gateway.yaml":     {"wait-for-aai-service"},
				"templates/aai-service.yaml": {"name: aai-service"},
				"templates/pvc.yaml":         {"name: aai"},
			},
		},
		{
			name: "monitoring enabled renders security key and monitoring settings",
			mutate: func(cfg *config.Config) {
				cfg.Name = "test-monitoring"
				cfg.Monitoring.Enabled = true
				cfg.Monitoring.URL = "https://monitoring.example.com"
				cfg.Monitoring.User = "monitor-user"
				cfg.Monitoring.Password = "monitor-password"
				cfg.Monitoring.SecurityKey = "test-security-key"
			},
			wantContains: map[string][]string{
				"templates/gateway.yaml": {
					`SECURITY_KEY: "test-security-key"`,
				},
				"templates/resources-service.yaml": {
					`MONITORING: "true"`,
					"MONITORING_URL: https://monitoring.example.com",
					"MONITORING_USER: monitor-user",
					"MONITORING_PWD: monitor-password",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.GetDefaultConfig()
			if tt.mutate != nil {
				tt.mutate(cfg)
			}

			files, err := cfg.Render()
			if err != nil {
				t.Fatalf("Render() returned error: %v", err)
			}

			for _, suffix := range tt.wantFiles {
				mustGetRenderedFileBySuffix(t, files, suffix)
			}

			for suffix, substrings := range tt.wantContains {
				content := mustGetRenderedFileBySuffix(t, files, suffix)
				contentContains(t, content, suffix, substrings)
			}

			for suffix, substrings := range tt.notContains {
				content, _ := getRenderedFileBySuffix(files, suffix)
				contentExcludes(t, content, suffix, substrings)
			}
		})
	}
}

func mustGetRenderedFileBySuffix(t *testing.T, files map[string]string, suffix string) string {
	t.Helper()

	content, ok := getRenderedFileBySuffix(files, suffix)
	if ok {
		return content
	}

	keys := make([]string, 0, len(files))
	for path := range files {
		keys = append(keys, path)
	}
	sort.Strings(keys)

	t.Fatalf("rendered file with suffix %q not found; got files: %v", suffix, keys)

	return ""
}

func getRenderedFileBySuffix(files map[string]string, suffix string) (string, bool) {
	for path, content := range files {
		if strings.HasSuffix(path, suffix) {
			return content, true
		}
	}

	return "", false
}

func contentContains(t *testing.T, content, fileName string, expected []string) {
	t.Helper()

	for _, s := range expected {
		if !strings.Contains(content, s) {
			t.Errorf("file %q missing expected content %q", fileName, s)
		}
	}
}

func contentExcludes(t *testing.T, content, fileName string, unexpected []string) {
	t.Helper()

	for _, s := range unexpected {
		if strings.Contains(content, s) {
			t.Errorf("file %q should not contain %q", fileName, s)
		}
	}
}
