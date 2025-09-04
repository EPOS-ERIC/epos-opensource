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
			opts:    UpdateOpts{Name: "test", CustomHost: "ht!tp://invalid-url"},
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
