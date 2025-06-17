package internal

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
				if got := mustRead(t, filepath.Join(envPath, ".env")); got != envFile {
					t.Fatalf(".env content mismatch")
				}

				// Each embedded manifest should have been written with its original filename and content.
				for fname, wantContent := range EmbeddedManifestContents {
					p := filepath.Join(envPath, fname)
					if got := mustRead(t, p); got != wantContent {
						t.Fatalf("%s content mismatch", fname)
					}
				}

				// No extra files besides .env and the manifests
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

				// create custom .env
				customEnvFile, _ := os.CreateTemp(base, "custom.env")
				wantEnvContent := "FOO=bar\n"
				customEnvFile.WriteString(wantEnvContent)
				customEnvFile.Close()

				// custom manifests dir
				manifDir := filepath.Join(base, "manifests")
				os.Mkdir(manifDir, 0700)
				os.WriteFile(filepath.Join(manifDir, "a.yaml"), []byte("a: 1\n"), 0600)
				os.WriteFile(filepath.Join(manifDir, "b.yaml"), []byte("b: 2\n"), 0600)

				return customEnvFile.Name(), manifDir, base, "custom", false
			},
			wantErr: false,
			check: func(t *testing.T, envPath, customEnv, customManifests string) {
				// .env content must match custom file
				if want, got := mustRead(t, customEnv), mustRead(t, filepath.Join(envPath, ".env")); want != got {
					t.Fatalf(".env mismatch: want %q got %q", want, got)
				}

				// manifest names and contents must match custom dir
				wantEntries, _ := os.ReadDir(customManifests)
				gotEntries, _ := os.ReadDir(envPath)

				var wantNames, gotNames []string
				for _, e := range wantEntries {
					wantNames = append(wantNames, e.Name())
					if got := mustRead(t, filepath.Join(envPath, e.Name())); got != mustRead(t, filepath.Join(customManifests, e.Name())) {
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			customEnv, customManif, basePath, envName, pre := tt.setup(t)

			if pre {
				os.MkdirAll(filepath.Join(basePath, envName), 0700)
			}

			envPath, err := NewEnvDir(customEnv, customManif, basePath, envName)
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
