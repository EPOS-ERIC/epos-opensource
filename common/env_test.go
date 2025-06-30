package common

import (
	"testing"

	"github.com/epos-eu/epos-opensource/db"
)

func TestGetEnvDir(t *testing.T) {
	t.Parallel()

	testDir := t.TempDir()
	const testPlatform = "docker"

	tests := []struct {
		name      string
		envName   string
		insertEnv bool
		expectErr bool
		expectDir string
	}{
		{
			name:      "existing environment returns correct dir",
			envName:   "testenvdir_env",
			insertEnv: true,
			expectErr: false,
			expectDir: testDir,
		},
		{
			name:      "non-existent environment returns error",
			envName:   "nonexistent_env",
			insertEnv: false,
			expectErr: true,
			expectDir: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.insertEnv {
				err := db.InsertEnv(tc.envName, testDir, testPlatform)
				if err != nil {
					t.Fatalf("failed to insert test env: %v", err)
				}
				defer func() { _ = db.DeleteEnv(tc.envName, testPlatform) }()
			}

			dir, err := GetEnvDir("", tc.envName, testPlatform)
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if dir != tc.expectDir {
					t.Errorf("got dir %q, want %q", dir, tc.expectDir)
				}
			}
		})
	}
}
