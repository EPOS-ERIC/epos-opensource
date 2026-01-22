package dockercore

import (
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore/config"
)

func TestUpdateOpts_Validate(t *testing.T) {
	invalidConfig := &config.EnvConfig{}

	tests := []struct {
		name    string
		opts    UpdateOpts
		wantErr bool
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
