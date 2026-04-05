package config_test

import (
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

func TestDockerEnvConfig_Render(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.EnvConfig
		wantErr      bool
		wantContains map[string][]string
		notContains  map[string][]string
	}{
		{
			name: "embedded default config renders core services with optional components disabled",
			config: func() *config.EnvConfig {
				cfg := config.GetDefaultConfig()
				cfg.Name = "default-env"
				return cfg
			}(),
			wantErr: false,
			wantContains: map[string][]string{
				".env":                {"ENV_NAME=default-env", "DATAPORTAL_PORT=32000", "GATEWAY_PORT=33000"},
				"docker-compose.yaml": {"dataportal:", "gateway:", "rabbitmq:", "metadata-database:"},
			},
			notContains: map[string][]string{
				".env":                {"LOAD_BACKOFFICE_API=", "LOAD_CONVERTER_API=", "LOAD_EMAIL_SENDER_API=", "LOAD_SHARING_API="},
				"docker-compose.yaml": {"backoffice-ui:", "backoffice-service:", "converter-service:", "converter-routine:", "email-sender-service:", "sharing-service:"},
			},
		},
		{
			name:    "backoffice enabled includes backoffice services in compose",
			config:  NewTestConfig(t, "test-bo").WithBackoffice(true).Build(),
			wantErr: false,
			wantContains: map[string][]string{
				".env":                {"ENV_NAME=test-bo"},
				"docker-compose.yaml": {"backoffice-ui:", "backoffice-service:"},
			},
		},
		{
			name: "backoffice uses embedded aai auth root url by default",
			config: func() *config.EnvConfig {
				cfg := NewTestConfig(t, "test-bo-aai").WithBackoffice(true).Build()
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.AAIService.Enabled = true
				return cfg
			}(),
			wantErr: false,
			wantContains: map[string][]string{
				".env":                {"AUTH_ROOT_URL=http://localhost:35000"},
				"docker-compose.yaml": {"dataportal:", "backoffice-ui:", "AUTH_ROOT_URL=${AUTH_ROOT_URL}"},
			},
		},
		{
			name:    "backoffice disabled excludes backoffice from compose",
			config:  NewTestConfig(t, "test-no-bo").Build(),
			wantErr: false,
			wantContains: map[string][]string{
				"docker-compose.yaml": {"dataportal:", "gateway:"},
			},
			notContains: map[string][]string{
				"docker-compose.yaml": {"backoffice-ui:", "backoffice-service:"},
			},
		},
		{
			name:    "custom ports are reflected in output",
			config:  NewTestConfig(t, "custom-ports").WithPorts(12345, 54321).Build(),
			wantErr: false,
			wantContains: map[string][]string{
				".env": {"DATAPORTAL_PORT=12345", "GATEWAY_PORT=54321"},
			},
		},
		{
			name:    "converter enabled includes converter services in compose",
			config:  NewTestConfig(t, "test-converter").WithConverter(true).Build(),
			wantErr: false,
			wantContains: map[string][]string{
				"docker-compose.yaml": {"converter-service:", "converter-routine:"},
			},
		},
		{
			name:    "sharing enabled includes sharing service in compose",
			config:  NewTestConfig(t, "test-sharing").WithSharing(true).Build(),
			wantErr: false,
			wantContains: map[string][]string{
				"docker-compose.yaml": {"sharing-service:"},
			},
		},
		{
			name: "embedded aai enabled includes aai service and env vars",
			config: func() *config.EnvConfig {
				cfg := NewTestConfig(t, "test-aai").Build()
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.AAIService.Enabled = true
				return cfg
			}(),
			wantErr: false,
			wantContains: map[string][]string{
				".env": {
					"IS_AAI_ENABLED=true",
					"AAI_SERVICE_ENDPOINT=http://aai-service:8080/oauth2/userinfo",
					"ADMIN_NAME=EPOS",
					"ADMIN_SURNAME=User",
					"ADMIN_EMAIL=epos@epos.eu",
					"ADMIN_PASSWORD=epos",
					"AAI_SERVICE_PORT=35000",
				},
				"docker-compose.yaml": {"aai-service:", "${AAI_SERVICE_PORT}:8080"},
			},
		},
		{
			name: "embedded aai disabled excludes service specific env vars and compose service",
			config: func() *config.EnvConfig {
				cfg := NewTestConfig(t, "test-external-aai").Build()
				cfg.Components.Gateway.AAI.Enabled = true
				cfg.Components.Gateway.AAI.ServiceEndpoint = externalAAIUserinfoEndpoint
				return cfg
			}(),
			wantErr: false,
			wantContains: map[string][]string{
				".env": {
					"IS_AAI_ENABLED=true",
					"AAI_SERVICE_ENDPOINT=" + externalAAIUserinfoEndpoint,
				},
			},
			notContains: map[string][]string{
				".env":                {"ADMIN_NAME=", "ADMIN_SURNAME=", "ADMIN_EMAIL=", "ADMIN_PASSWORD=", "AAI_SERVICE_PORT="},
				"docker-compose.yaml": {"aai-service:"},
			},
		},
		{
			name: "metadata database published port is rendered when configured",
			config: func() *config.EnvConfig {
				cfg := NewTestConfig(t, "test-db-port").Build()
				cfg.Components.MetadataDatabase.PublishedPort = 35432
				return cfg
			}(),
			wantErr: false,
			wantContains: map[string][]string{
				".env":                {"METADATA_DATABASE_PUBLISHED_PORT=35432"},
				"docker-compose.yaml": {"\"${METADATA_DATABASE_PUBLISHED_PORT}:5432\""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.config.Render()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Render() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Render() succeeded unexpectedly")
			}

			for _, key := range []string{"docker-compose.yaml", ".env"} {
				if _, ok := got[key]; !ok {
					t.Errorf("Render() missing key %q", key)
				}
			}

			for file, substrings := range tt.wantContains {
				content := got[file]
				ContentContains(t, content, file, substrings)
			}

			for file, substrings := range tt.notContains {
				content := got[file]
				ContentExcludes(t, content, file, substrings)
			}
		})
	}
}

func TestDockerEnvConfig_Render_EmailSenderAuth(t *testing.T) {
	cfg := NewTestConfig(t, "test-email-auth").WithEmailSender(true).Build()
	cfg.Components.EmailSenderService.Auth = config.Auth{
		Enabled:   true,
		OnlyAdmin: true,
	}

	got := MustRender(t, cfg)
	ContentContains(t, got[".env"], ".env", []string{"LOAD_EMAIL_SENDER_API=true:true"})
}
