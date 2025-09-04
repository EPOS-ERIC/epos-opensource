package dockercore

import (
	"testing"

	"github.com/epos-eu/epos-opensource/db/sqlc"
)

func TestUpdate(t *testing.T) {
	t.Helper()

	type setupFn func(t *testing.T) (opts UpdateOpts, cleanup func())

	tests := []struct {
		name    string
		setup   setupFn
		wantErr bool
		check   func(t *testing.T, got *sqlc.Docker)
	}{
		{
			name: "failure – non-existent env",
			setup: func(t *testing.T) (UpdateOpts, func()) {
				return UpdateOpts{Name: "does-not-exist"}, func() {}
			},
			wantErr: true,
		},
		{
			name: "success simple update",
			setup: func(t *testing.T) (UpdateOpts, func()) {
				base := "update-basic"
				_, err := Deploy(DeployOpts{Name: base})
				if err != nil {
					t.Fatalf("deploy failed: %v", err)
				}
				return UpdateOpts{Name: base}, func() {
					_ = Delete(DeleteOpts{Name: base})
				}
			},
			wantErr: false,
			check: func(t *testing.T, got *sqlc.Docker) {
				if got.ApiUrl == "" || got.GuiUrl == "" || got.BackofficeUrl == "" {
					t.Errorf("urls should not be empty: %+v", got)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			opts, cleanup := tt.setup(t)
			defer cleanup()

			got, err := Update(opts)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("Update() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err != nil {
				return
			}

			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
