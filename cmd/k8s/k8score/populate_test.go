package k8score

import (
	"testing"
)

func TestPopulateOpts_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    PopulateOpts
		wantErr bool
	}{
		{
			name:    "Environment does not exist",
			opts:    PopulateOpts{Name: "does-not-exist"},
			wantErr: true,
		},
		{
			name:    "To many files in parallel",
			opts:    PopulateOpts{Name: "test", Parallel: 21},
			wantErr: true,
		},
		{
			name:    "Negative number in parallel",
			opts:    PopulateOpts{Name: "test", Parallel: -1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("Populate K8s Validation Test Failed error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
