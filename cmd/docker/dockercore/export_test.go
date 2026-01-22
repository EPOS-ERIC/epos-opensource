package dockercore

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore/config"
)

func TestExport(t *testing.T) {
	tests := []struct {
		name          string
		setupPath     func(t *testing.T) (string, func())
		wantErr       bool
		wantFileMatch bool
	}{
		{
			name: "success - existing dir",
			setupPath: func(t *testing.T) (string, func()) {
				return t.TempDir(), func() {}
			},
			wantErr:       false,
			wantFileMatch: true,
		},
		{
			name: "success - dir auto-created",
			setupPath: func(t *testing.T) (string, func()) {
				path := filepath.Join(t.TempDir(), "nested", "dir")
				return path, func() {}
			},
			wantErr:       false,
			wantFileMatch: true,
		},
		{
			name: "failure - path is a file",
			setupPath: func(t *testing.T) (string, func()) {
				tmpFile, err := os.CreateTemp(t.TempDir(), "not-a-dir")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}
				_ = tmpFile.Close()
				return tmpFile.Name(), func() {}
			},
			wantErr:       true,
			wantFileMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exportPath, cleanup := tt.setupPath(t)
			defer cleanup()

			err := Export(ExportOpts{Path: exportPath})

			if (err != nil) != tt.wantErr {
				t.Fatalf("Export() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				filePath := filepath.Join(exportPath, "config.yaml")
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("Export() did not create config.yaml")
				}

				if tt.wantFileMatch {
					content, err := os.ReadFile(filePath)
					if err != nil {
						t.Errorf("failed to read exported file: %v", err)
					}
					expected := config.GetDefaultConfigBytes()
					if !bytes.HasSuffix(content, expected) {
						t.Errorf("Export() content does not end with default config")
					}
				}
			}
		})
	}
}
