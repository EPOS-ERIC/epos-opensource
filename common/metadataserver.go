package common

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

// PostFiles walks the server's directory tree and POSTs **every file
// ending in `.ttl`** to the EPOS gateway associated with *gatewayURL*.
//
//   - gatewayURL should be the base URL of the EPOS Gateway (e.g.
//     "http://localhost:8080" or "https://gateway.epos.eu").  The
//     function will automatically append the "/populate" endpoint and
//     encode the required query parameters.
//   - The MetadataServer **must be running**; the function uses Addr()
//     to build public URLs for each TTL file.
//   - If any file fails to ingest—or if directory traversal itself
//     fails—the function returns a non‑nil error.  The ingestion stops
//     at the first fatal directory walk error but continues on HTTP
//     errors.
func (ms *MetadataServer) PostFiles(gatewayURL string) error {
	postURL, err := url.Parse(gatewayURL)
	if err != nil {
		return fmt.Errorf("error parsing url '%s': %w", gatewayURL, err)
	}
	postURL = postURL.JoinPath("/populate")

	PrintDone("Deployed metadata server with mounted dir: %s", ms.dir)
	PrintStep("Starting the ingestion of the '*.ttl' files")

	ingestionError := false
	err = filepath.WalkDir(ms.dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			PrintError("Error while walking directory: %v", walkErr)
			ingestionError = true
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".ttl") {
			return nil
		}

		PrintStep("Ingesting file: %s", d.Name())
		relPath, err := filepath.Rel(ms.dir, path)
		if err != nil {
			PrintError("Error getting relative path: %v", err)
			ingestionError = true
			return nil
		}

		q := postURL.Query()
		// TODO: remove securityCode once it's removed from the ingestor
		q.Set("securityCode", "changeme")
		q.Set("type", "single")
		q.Set("model", "EPOS-DCAT-AP-V1")
		q.Set("mapping", "EDM-TO-DCAT-AP")

		fileURL, err := url.JoinPath("http://", ms.Addr(), relPath)
		if err != nil {
			PrintError("Error while building URL for file '%s': %v", d.Name(), err)
			ingestionError = true
			return nil
		}

		q.Set("path", fileURL)
		postURL.RawQuery = q.Encode()

		r, err := http.NewRequest("POST", postURL.String(), nil)
		if err != nil {
			PrintError("Error building request for file '%s': %v", d.Name(), err)
			ingestionError = true
			return nil
		}

		r.Header.Add("accept", "*/*")
		res, err := http.DefaultClient.Do(r)
		if err != nil {
			PrintError("Error ingesting file '%s' in database: %v", d.Name(), err)
			ingestionError = true
			return nil
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			PrintError("Error reading response body for file '%s': %v", d.Name(), err)
			ingestionError = true
			return nil
		}

		if res.StatusCode != http.StatusOK {
			PrintError("Error ingesting file '%s' in database: received status code %d. Body of response: %s", d.Name(), res.StatusCode, string(body))
			ingestionError = true
			return nil
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to ingest metadata in directory %s: %w", ms.dir, err)
	}

	if ingestionError {
		return fmt.Errorf("failed to ingest metadata in directory %s", ms.dir)
	}
	return nil
}
