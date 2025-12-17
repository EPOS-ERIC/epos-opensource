package common

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestPopulateEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		files         map[string]string // path -> content
		targetPath    string            // path to pass to PopulateEnv
		serverHandler http.HandlerFunc
		expectErr     bool
		expectedPaths []string // paths expected at the server
	}{
		{
			name: "single_file_success",
			files: map[string]string{
				"single.ttl": "file content",
			},
			targetPath: "single.ttl",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if body, _ := io.ReadAll(r.Body); string(body) != "file content" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
			expectErr:     false,
			expectedPaths: []string{"/populate"},
		},
		{
			name: "directory_success",
			files: map[string]string{
				"a.ttl":          "content a",
				"b.ttl":          "content b",
				"c.txt":          "content c",
				"sub/d.ttl":      "content d",
				"sub/e.json":     "content e",
				"sub/sub2/f.ttl": "content f",
			},
			targetPath: ".",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectErr: false,
			expectedPaths: []string{
				"/populate",
				"/populate",
				"/populate",
				"/populate",
			},
		},
		{
			name: "server_error",
			files: map[string]string{
				"bad.ttl": "this will fail",
			},
			targetPath: "bad.ttl",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectErr:     true,
			expectedPaths: []string{"/populate"},
		},
		{
			name:          "file_not_found",
			files:         nil,
			targetPath:    "nonexistent.ttl",
			serverHandler: nil,
			expectErr:     true,
			expectedPaths: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory for the test files
			tmpDir := t.TempDir()

			// Create the files for the current test case
			for path, content := range tc.files {
				fullPath := filepath.Join(tmpDir, path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
					t.Fatalf("Failed to write file: %v", err)
				}
			}

			// Mock Server
			var serverURL string
			var receivedPaths []string
			var mu sync.Mutex // Mutex to protect receivedPaths

			if tc.serverHandler != nil {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mu.Lock()
					receivedPaths = append(receivedPaths, r.URL.Path)
					mu.Unlock()
					tc.serverHandler(w, r)
				}))
				defer server.Close()
				serverURL = server.URL
			}

			target := filepath.Join(tmpDir, tc.targetPath)
			_, err := PopulateEnv(target, serverURL, 2)

			if tc.expectErr {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}

			// Check if the received paths at the server match the expected paths
			if len(receivedPaths) != len(tc.expectedPaths) {
				t.Errorf("Expected %d requests, but got %d", len(tc.expectedPaths), len(receivedPaths))
			}
		})
	}
}
