package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveStartPath(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(tempFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		startPath string
		expected  string
	}{
		{
			name:      "empty startPath uses CWD",
			startPath: "",
			expected: func() string {
				cwd, _ := os.Getwd()
				abs, _ := filepath.Abs(cwd)
				return abs
			}(),
		},
		{
			name:      "startPath is file resolves to dir",
			startPath: tempFile,
			expected:  tempDir,
		},
		{
			name:      "startPath is dir no change",
			startPath: tempDir,
			expected:  tempDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveStartPath(tt.startPath)
			if result != tt.expected {
				t.Errorf("resolveStartPath(%q) = %q; want %q", tt.startPath, result, tt.expected)
			}
		})
	}
}
