package config_test

import (
	"reflect"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/common"
)

func TestEnvConfig_ActiveImages(t *testing.T) {
	tests := []struct {
		name string
		cfg  *TestConfigBuilder
		want []common.NamedImage
	}{
		{
			name: "base config includes core and enabled sharing images",
			cfg:  NewTestConfig(t, "test"),
			want: []common.NamedImage{
				{Name: "Rabbitmq", Ref: "rabbitmq:3.13.7-management"},
				{Name: "Platform UI", Ref: "epos/data-portal:latest"},
				{Name: "Gateway", Ref: "ghcr.io/epos-eric/epos-api-gateway:latest"},
				{Name: "Metadata Database", Ref: "ghcr.io/epos-eric/metadata-database/deploy:latest"},
				{Name: "Resources Service", Ref: "ghcr.io/epos-eric/resources-service:latest"},
				{Name: "Ingestor Service", Ref: "ghcr.io/epos-eric/ingestor-service:latest"},
				{Name: "External Access Service", Ref: "ghcr.io/epos-eric/external-access-service:latest"},
				{Name: "Sharing Service", Ref: "ghcr.io/epos-eric/sharing-service:latest"},
			},
		},
		{
			name: "optional components add their images in a stable order",
			cfg: NewTestConfig(t, "test").
				WithConverter(true).
				WithBackoffice(true).
				WithEmailSender(true).
				WithSharing(false),
			want: []common.NamedImage{
				{Name: "Rabbitmq", Ref: "rabbitmq:3.13.7-management"},
				{Name: "Platform UI", Ref: "epos/data-portal:latest"},
				{Name: "Gateway", Ref: "ghcr.io/epos-eric/epos-api-gateway:latest"},
				{Name: "Metadata Database", Ref: "ghcr.io/epos-eric/metadata-database/deploy:latest"},
				{Name: "Resources Service", Ref: "ghcr.io/epos-eric/resources-service:latest"},
				{Name: "Ingestor Service", Ref: "ghcr.io/epos-eric/ingestor-service:latest"},
				{Name: "External Access Service", Ref: "ghcr.io/epos-eric/external-access-service:latest"},
				{Name: "Converter Service", Ref: "ghcr.io/epos-eric/converter-service-go:latest"},
				{Name: "Converter Routine", Ref: "ghcr.io/epos-eric/converter-routine-go:latest"},
				{Name: "Backoffice Service", Ref: "ghcr.io/epos-eric/backoffice-service:latest"},
				{Name: "Backoffice UI", Ref: "epos/backoffice-ui:latest"},
				{Name: "Email Sender Service", Ref: "ghcr.io/epos-eric/email-sender-service:latest"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.Build().ActiveImages()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ActiveImages() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
