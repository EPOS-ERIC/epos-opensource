package common

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type MetadataServer struct {
	dir  string       // absolute path served
	Srv  *http.Server // underlying HTTP server
	addr string       // full address "ip:port", e.g. "192.168.1.20:53513"
}

// NewMetadataServer validates ttlDir and prepares the object.
// The listener is created when Start() is called.
func NewMetadataServer(ttlDir string) (*MetadataServer, error) {
	absDir, err := filepath.Abs(ttlDir)
	if err != nil {
		return nil, fmt.Errorf("resolve ttlDir: %w", err)
	}
	if fi, err := os.Stat(absDir); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("ttlDir %q is not a directory: %w", absDir, err)
	}
	return &MetadataServer{dir: absDir}, nil
}

// Start binds an ephemeral port on all interfaces, launches the HTTP
// server in a goroutine, and records "<local-ip>:<port>".
func (ms *MetadataServer) Start() error {
	ln, err := net.Listen("tcp", ":0") // 0.0.0.0, OS picks port
	if err != nil {
		return fmt.Errorf("listen on ephemeral port: %w", err)
	}

	// Determine the primary non-loopback IPv4 of the host.
	ip, err := GetLocalIP()
	if err != nil {
		_ = ln.Close()
		return err
	}

	port := ln.Addr().(*net.TCPAddr).Port
	ms.addr = net.JoinHostPort(ip, fmt.Sprint(port))

	ms.Srv = &http.Server{
		Handler: http.FileServer(http.Dir(ms.dir)),
	}

	go func() {
		_ = ms.Srv.Serve(ln) // errors handled via Stop
	}()
	return nil
}

// Stop performs a graceful shutdown with a 5-second timeout.
func (ms *MetadataServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ms.Srv.Shutdown(ctx)
}

// Addr returns the full "ip:port" string (valid after Start()).
func (ms *MetadataServer) Addr() string { return ms.addr }
