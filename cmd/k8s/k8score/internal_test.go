package k8score

import (
	"os"
	"path/filepath"
	"testing"
)

// helper: read an entire file or die
func mustRead(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("reading %s: %v", p, err)
	}
	return string(b)
}

func TestNewEnvDir(t *testing.T) {
	type tc struct {
		name    string
		setup   func(t *testing.T) (customEnv, customManifests, basePath, envName string, preCreate bool)
		wantErr bool
		check   func(t *testing.T, envPath, customEnv, customManifests string)
	}

	tests := []tc{
		{
			name: "default embedded assets",
			setup: func(t *testing.T) (string, string, string, string, bool) {
				return "", "", t.TempDir(), "default", false
			},
			wantErr: false,
			check: func(t *testing.T, envPath, _, _ string) {
				// .env should equal the embedded envFile
				_ = os.Setenv("NAMESPACE", "default")
				expandedEnvFile := os.ExpandEnv(EnvFile)
				if got := mustRead(t, filepath.Join(envPath, ".env")); got != expandedEnvFile {
					t.Fatalf(".env content mismatch. got=\n%s, want=\n%s", got, expandedEnvFile)
				}

				// Each embedded manifest should be present with its ENV-expanded content.
				for fname, raw := range EmbeddedManifestContents {
					got := mustRead(t, filepath.Join(envPath, fname))
					want := os.ExpandEnv(raw)
					if want != got {
						t.Fatalf("%s content mismatch:\nwant:\n%s\ngot:\n%s", fname, want, got)
					}
				}

				// No extra files besides .env and the manifests.
				files, _ := os.ReadDir(envPath)
				wantFileCount := 1 + len(EmbeddedManifestContents)
				if len(files) != wantFileCount {
					t.Fatalf("expected %d files, got %d", wantFileCount, len(files))
				}
			},
		},
		{
			name: "directory already exists",
			setup: func(t *testing.T) (string, string, string, string, bool) {
				base := t.TempDir()
				return "", "", base, "exists", true
			},
			wantErr: true,
		},
		{
			name: "bad custom env path",
			setup: func(t *testing.T) (string, string, string, string, bool) {
				return filepath.Join(t.TempDir(), "nowhere.env"), "", t.TempDir(), "badenv", false
			},
			wantErr: true,
			check: func(t *testing.T, envPath, _, _ string) {
				if _, err := os.Stat(envPath); !os.IsNotExist(err) {
					t.Fatalf("expected cleanup of env dir on error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customEnv, customManif, basePath, envName, preCreate := tt.setup(t)

			if preCreate {
				_ = os.MkdirAll(filepath.Join(basePath, envName), 0o700)
			}

			envPath, err := NewEnvDir(customEnv, customManif, basePath, envName, "docker-desktop", "http", "localhost")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.check != nil {
					tt.check(t, filepath.Join(basePath, envName), customEnv, customManif)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			t.Cleanup(func() { _ = os.RemoveAll(envPath) })

			if tt.check != nil {
				tt.check(t, envPath, customEnv, customManif)
			}
		})
	}
}
