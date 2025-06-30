package k8score

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
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
				if got := mustRead(t, filepath.Join(envPath, ".env")); got != EnvFile {
					t.Fatalf(".env content mismatch. got=\n%s, want=\n%s", got, EnvFile)
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
			name: "custom env & manifests",
			setup: func(t *testing.T) (string, string, string, string, bool) {
				base := t.TempDir()

				// custom .env
				customEnvFile, _ := os.CreateTemp(base, "custom.env")
				customEnvContent := "FOO=bar\nAPI_PATH=something"
				customEnvFile.WriteString(customEnvContent)
				customEnvFile.Close()

				// custom manifests dir
				manifDir := filepath.Join(base, "manifests")
				os.Mkdir(manifDir, 0o700)
				os.WriteFile(filepath.Join(manifDir, "a.yaml"), []byte("a: 1\n"), 0o600)
				os.WriteFile(filepath.Join(manifDir, "b.yaml"), []byte("b: 2\n"), 0o600)

				return customEnvFile.Name(), manifDir, base, "custom", false
			},
			wantErr: false,
			check: func(t *testing.T, envPath, customEnv, customManifests string) {
				// .env must match the custom file
				if want, got := mustRead(t, customEnv), mustRead(t, filepath.Join(envPath, ".env")); want != got {
					t.Fatalf(".env mismatch: want %q got %q", want, got)
				}

				// Manifest names and contents must match the custom dir (after expansion,
				// which is a no-op for these simple files).
				wantEntries, _ := os.ReadDir(customManifests)
				gotEntries, _ := os.ReadDir(envPath)

				var wantNames, gotNames []string
				for _, e := range wantEntries {
					wantNames = append(wantNames, e.Name())
					wantContent := mustRead(t, filepath.Join(customManifests, e.Name()))
					gotContent := mustRead(t, filepath.Join(envPath, e.Name()))
					if wantContent != gotContent {
						t.Fatalf("content mismatch in %s", e.Name())
					}
				}
				for _, e := range gotEntries {
					if e.Name() != ".env" {
						gotNames = append(gotNames, e.Name())
					}
				}
				sort.Strings(wantNames)
				sort.Strings(gotNames)
				if !reflect.DeepEqual(wantNames, gotNames) {
					t.Fatalf("manifest name set mismatch: want %v got %v", wantNames, gotNames)
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
				os.MkdirAll(filepath.Join(basePath, envName), 0o700)
			}

			envPath, err := NewEnvDir(customEnv, customManif, basePath, envName, "docker-desktop")
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
			t.Cleanup(func() { os.RemoveAll(envPath) })

			if tt.check != nil {
				tt.check(t, envPath, customEnv, customManif)
			}
		})
	}
}
