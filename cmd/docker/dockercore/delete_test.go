package dockercore

import (
	"testing"
)

func TestDeleteOpts_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    DeleteOpts
		wantErr bool
	}{
		{
			name:    "non-existent environment returns error",
			opts:    DeleteOpts{Name: []string{"does_not_exist"}},
			wantErr: true,
		},
		{
			name:    "empty slice returns no error",
			opts:    DeleteOpts{Name: []string{}},
			wantErr: false,
		},
		{
			name:    "first of multiple non-existent environments fails",
			opts:    DeleteOpts{Name: []string{"env1", "does_not_exist", "env3"}},
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
