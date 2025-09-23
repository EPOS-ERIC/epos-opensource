package metadataserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// waitAddrReady dials addr ("ip:port") until a connection succeeds or
// the deadline expires.
func waitAddrReady(addr string, d time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	for {
		var dialer net.Dialer
		if conn, err := dialer.DialContext(ctx, "tcp", addr); err == nil {
			err := conn.Close()
			if err != nil {
				return fmt.Errorf("failed to close the connection: %w", err)
			}
			return nil // listener is up
		}
		if ctx.Err() != nil {
			return fmt.Errorf("addr %s not ready within %v", addr, d)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestMetadataServer(t *testing.T) {
	t.Parallel() // safe: implementation no longer mutates global state

	tests := []struct {
		name         string
		makeDir      bool   // create a real temp dir if true
		expectErrNew bool   // expect NewMetadataServer to error
		wantBody     string // expected response body (for success cases)
	}{
		{
			name:     "serves static file from valid dir",
			makeDir:  true,
			wantBody: "hello world\n",
		},
		{
			name:         "returns error for missing dir",
			makeDir:      false,
			expectErrNew: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := "/path/does/not/exist"
			if tc.makeDir {
				dir = t.TempDir()
				// create a sample file the server should serve
				if err := os.WriteFile(filepath.Join(dir, "greet.txt"), []byte(tc.wantBody), 0o600); err != nil {
					t.Fatalf("write fixture: %v", err)
				}
			}

			ms, err := New(dir, 2)
			if tc.expectErrNew {
				if err == nil {
					t.Fatalf("expected error from NewMetadataServer, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewMetadataServer() error = %v", err)
			}

			// Start the server.
			if err := ms.Start(); err != nil {
				t.Fatalf("Start() error = %v", err)
			}
			t.Cleanup(func() {
				_ = ms.Stop() // best-effort shutdown
			})

			// Basic sanity: Addr should split into host + numeric port
			host, portStr, err := net.SplitHostPort(ms.Addr())
			if err != nil || host == "" {
				t.Fatalf("Addr() returned malformed value %q", ms.Addr())
			}
			if _, err := strconv.Atoi(portStr); err != nil {
				t.Fatalf("Addr() returned non-numeric port %q", ms.Addr())
			}

			// Wait until the listener is accepting connections.
			if err := waitAddrReady(ms.Addr(), time.Second); err != nil {
				t.Fatalf("server never became reachable: %v", err)
			}

			// Fetch the fixture file and compare.
			resp, err := http.Get(fmt.Sprintf("http://%s/greet.txt", ms.Addr()))
			if err != nil {
				t.Fatalf("HTTP GET failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("unexpected status %d", resp.StatusCode)
			}
			got, _ := io.ReadAll(resp.Body)
			if string(got) != tc.wantBody {
				t.Errorf("body = %q, want %q", got, tc.wantBody)
			}
		})
	}
}
