package config_test

import (
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore/config"
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
			name:    "default config renders both files successfully",
			config:  NewTestConfig(t, "test-env").Build(),
			wantErr: false,
			wantContains: map[string][]string{
				".env":                {"ENV_NAME=test-env", "DATAPORTAL_PORT=32000", "GATEWAY_PORT=33000"},
				"docker-compose.yaml": {"dataportal:", "gateway:", "rabbitmq:", "metadata-database:"},
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
			name:    "sharing service appears in compose",
			config:  NewTestConfig(t, "test-sharing").Build(),
			wantErr: false,
			wantContains: map[string][]string{
				"docker-compose.yaml": {"sharing-service:"},
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
