package config_test

import (
	"fmt"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore/config"
)

func TestDockerEnvConfig_Render(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "test",
			want:    map[string]string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			d := config.GetDefaultConfig()
			got, gotErr := d.Render()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Render() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Render() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			fmt.Printf("got: %v\n", got)
		})
	}
}
