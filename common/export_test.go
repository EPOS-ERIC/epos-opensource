package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExport(t *testing.T) {
	t.Helper()

	type setupFn func(t *testing.T) (path string, cleanup func())

	mkContent := func() []byte { return []byte("hello world") }

	tests := []struct {
		name     string
		setup    setupFn
		filename string
		content  []byte
		wantErr  bool
	}{
		{
			name: "success – existing dir",
			setup: func(t *testing.T) (string, func()) {
				return t.TempDir(), func() {}
			},
			filename: "file.txt",
			content:  mkContent(),
			wantErr:  false,
		},
		{
			name: "success – dir auto-created",
			setup: func(t *testing.T) (string, func()) {
				parent := t.TempDir()
				// subPath does not exist yet; Export should create it.
				return filepath.Join(parent, "nested", "deeper"), func() {}
			},
			filename: "file.txt",
			content:  mkContent(),
			wantErr:  false,
		},
		{
			name: "success – empty path uses cwd",
			setup: func(t *testing.T) (string, func()) {
				tmp := t.TempDir()
				orig, _ := os.Getwd()

				if err := os.Chdir(tmp); err != nil {
					t.Fatalf("chdir: %v", err)
				}
				return "", func() { _ = os.Chdir(orig) }
			},
			filename: "file.txt",
			content:  mkContent(),
			wantErr:  false,
		},
		{
			name: "failure – path is a file, not a directory",
			setup: func(t *testing.T) (string, func()) {
				parent := t.TempDir()
				f := filepath.Join(parent, "notADir")
				if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
					t.Fatalf("prep file: %v", err)
				}
				return f, func() {}
			},
			filename: "file.txt",
			content:  mkContent(),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path, cleanup := tt.setup(t)
			defer cleanup()

			err := Export(path, tt.filename, tt.content)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Export() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return // nothing further to verify on expected error
			}

			// Verify file content and perms for successful cases.
			fp := filepath.Join(
				func() string {
					if path == "" {
						wd, _ := os.Getwd()
						return wd
					}
					return path
				}(), tt.filename)

			gotData, err := os.ReadFile(fp)
			if err != nil {
				t.Fatalf("reading exported file: %v", err)
			}
			if !stringHasSuffix(string(gotData), string(tt.content)) {
				t.Errorf("file content does not end with expected content")
			}

			info, err := os.Stat(fp)
			if err != nil {
				t.Fatalf("stat exported file: %v", err)
			}
			if info.Mode().Perm() != 0o644 {
				t.Errorf("file perms %o, want 0o644", info.Mode().Perm())
			}

			// If Export had to create the directory, confirm it exists.
			dir := filepath.Dir(fp)
			if _, err := os.Stat(dir); err != nil {
				t.Fatalf("directory %s does not exist after Export: %v", dir, err)
			}
		})
	}
}

func stringHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
