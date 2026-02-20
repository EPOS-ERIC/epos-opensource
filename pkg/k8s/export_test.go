package k8s

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
)

func TestExport(t *testing.T) {
	tests := []struct {
		name      string
		setupPath func(t *testing.T) string
		wantErr   bool
	}{
		{
			name: "success - exports k8s config filename",
			setupPath: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: false,
		},
		{
			name: "failure - path is a file",
			setupPath: func(t *testing.T) string {
				tmpFile, err := os.CreateTemp(t.TempDir(), "not-a-dir")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}
				_ = tmpFile.Close()
				return tmpFile.Name()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exportPath := tt.setupPath(t)

			err := Export(ExportOpts{Path: exportPath})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Export() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			k8sConfigPath := filepath.Join(exportPath, "k8s-config.yaml")
			if _, err := os.Stat(k8sConfigPath); err != nil {
				t.Fatalf("Export() did not create k8s-config.yaml: %v", err)
			}

			legacyPath := filepath.Join(exportPath, "config.yaml")
			if _, err := os.Stat(legacyPath); err == nil {
				t.Fatalf("Export() unexpectedly created legacy config.yaml")
			} else if !os.IsNotExist(err) {
				t.Fatalf("stat legacy config.yaml: %v", err)
			}

			content, err := os.ReadFile(k8sConfigPath)
			if err != nil {
				t.Fatalf("failed to read exported file: %v", err)
			}

			expected := config.GetDefaultConfigBytes()
			if !bytes.HasSuffix(content, expected) {
				t.Fatalf("Export() content does not end with default k8s config")
			}
		})
	}
}
