package dockercore

import (
	_ "embed"
	"testing"
)

func TestDeleteOpts_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    DeleteOpts
		wantErr bool
	}{
		{
			name:    "Environment does not exist",
			opts:    DeleteOpts{Name: []string{"does_not_exist"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("DeleteOpts.Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
