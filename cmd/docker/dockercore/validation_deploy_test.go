package dockercore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore/config"
)

func TestDeployOpts_Validate(t *testing.T) {
	tmpDir := t.TempDir()

	validConfig := newTestConfig(t, "test-deploy")
	invalidConfig := &config.EnvConfig{}
	invalidNameConfig := newTestConfig(t, "787/(//&&$%£$£$£&)(/((=??=)))")

	tests := []struct {
		name    string
		opts    DeployOpts
		wantErr bool
	}{
		{
			name:    "Nil config returns error",
			opts:    DeployOpts{Config: nil},
			wantErr: true,
		},
		{
			name:    "Invalid config returns error",
			opts:    DeployOpts{Config: invalidConfig},
			wantErr: true,
		},
		{
			name:    "Invalid name returns error",
			opts:    DeployOpts{Config: invalidNameConfig},
			wantErr: true,
		},
		{
			name:    "Invalid path returns error",
			opts:    DeployOpts{Config: validConfig, Path: "/invalid/path"},
			wantErr: true,
		},
		{
			name:    "Valid opts with empty path",
			opts:    DeployOpts{Config: validConfig, Path: ""},
			wantErr: false,
		},
		{
			name:    "Valid opts with existing path",
			opts:    DeployOpts{Config: validConfig, Path: tmpDir},
			wantErr: false,
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

func TestDeployOpts_Validate_PathExists(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(filePath, []byte{}, 0o600); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	validConfig := newTestConfig(t, "test-validate")

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Empty path is valid",
			path:    "",
			wantErr: false,
		},
		{
			name:    "Existing directory is valid",
			path:    tmpDir,
			wantErr: false,
		},
		{
			name:    "Non-existent path is invalid",
			path:    "/this/path/does/not/exist",
			wantErr: true,
		},
		{
			name:    "Path is a file is invalid",
			path:    filePath,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DeployOpts{Config: validConfig, Path: tt.path}
			err := opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
