package dockercore

import (
	_ "embed"
	"testing"
)

func TestDelete(t *testing.T) {
	type setupFn func(t *testing.T) (opts DeleteOpts, cleanup func())
	tests := []struct {
		name    string
		setup   setupFn
		wantErr bool
	}{
		{
			name: "success: simple delete",
			setup: func(t *testing.T) (DeleteOpts, func()) {
				base := "delete-me"
				_, err := Deploy(DeployOpts{Name: base})
				if err != nil {
					t.Fatalf("deploy failed: %v", err)
				}
				return DeleteOpts{Name: base}, func() {
					_ = Delete(DeleteOpts{Name: base})
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, cleanup := tt.setup(t)
			defer cleanup()

			if err := Delete(opts); err != nil {
				if !tt.wantErr {
					t.Fatalf("success: simple delete test Failed error = %v", err)
				}
			}
		})
	}
}

func TestDeleteOpts_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    DeleteOpts
		wantErr bool
	}{
		{
			name:    "Environment does not exist",
			opts:    DeleteOpts{Name: "does-not-exist"},
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
