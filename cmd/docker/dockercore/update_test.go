package dockercore

import (
	"testing"
)

func TestUpdateOpts_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    UpdateOpts
		wantErr bool
	}{
		{
			name:    "Environment does not exist",
			opts:    UpdateOpts{Name: "does-not-exist"},
			wantErr: true,
		},
		{
			name:    "To many files in parallel",
			opts:    UpdateOpts{Name: "test"},
			wantErr: true,
		},
		{
			name:    "Reset with custom env file",
			opts:    UpdateOpts{Name: "test", Reset: true, EnvFile: "/fake/path"},
			wantErr: true,
		},
		{
			name:    "Reset with custom compose file",
			opts:    UpdateOpts{Name: "test", Reset: true, ComposeFile: "/fake/path"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("Update Validation Test Failed error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
