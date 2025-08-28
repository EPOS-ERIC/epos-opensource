// Package metadataserver spins up a tiny HTTP file server that exposes
// a directory full of TTL files and provides a helper to POST those
// files to an EPOS Gateway for ingestion.
//
// # Example
//
//	ms, _ := metadataserver.NewMetadataServer("/path/to/ttl")
//	_ = ms.Start()
//	defer ms.Stop()
//
//	// Send all *.ttl files to the gateway
//	_ = ms.PostFiles("https://gateway.epos.eu", "http")
package metadataserver

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

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"
)

type MetadataServer struct {
	dir      string       // absolute path served
	Srv      *http.Server // underlying HTTP server
	addr     string
	onlyFile string // full address "ip:port", e.g. "192.168.1.20:53513"
}

func (ms *MetadataServer) LimitToFile(file string) {
	ms.dir = filepath.Dir(file)       // serve parent dir
	ms.onlyFile = filepath.Base(file) // only expose this file
}

// NewMetadataServer validates ttlDir and prepares the object.
// The listener is created when Start() is called.
func NewMetadataServer(ttlDir string) (*MetadataServer, error) {
	absPath, err := filepath.Abs(ttlDir)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	fi, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("stat path %q: %w", absPath, err)
	}

	ms := &MetadataServer{}
	if fi.IsDir() {
		ms.dir = absPath
	} else {
		ms.LimitToFile(absPath)
	}

	return ms, nil
}

// Start binds an ephemeral port on all interfaces, launches the HTTP
// server in a goroutine, and records "<local-ip>:<port>".
func (ms *MetadataServer) Start() error {
	ln, err := net.Listen("tcp", ":0") // 0.0.0.0, OS picks port
	if err != nil {
		return fmt.Errorf("listen on ephemeral port: %w", err)
	}

	// Determine the primary non-loopback IPv4 of the host.
	ip, err := common.GetLocalIP()
	if err != nil {
		_ = ln.Close()
		return err
	}

	port := ln.Addr().(*net.TCPAddr).Port
	ms.addr = net.JoinHostPort(ip, fmt.Sprint(port))

	ms.Srv = &http.Server{
		ReadHeaderTimeout: 15 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If LimitToFile was called, only serve that file
			if ms.onlyFile != "" && filepath.Base(r.URL.Path) != ms.onlyFile {
				http.NotFound(w, r)
				return
			}
			// Serve the requested file
			http.FileServer(http.Dir(ms.dir)).ServeHTTP(w, r)
		}),
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

// PostFiles walks the server's directory tree and POSTs every file
// ending in .ttl to the EPOS gateway associated with gatewayURL.
//
//   - gatewayURL should be the base URL of the EPOS Gateway (e.g.
//     "http://localhost:8080" or "https://gateway.epos.eu").  The
//     function will automatically append the "/populate" endpoint and
//     encode the required query parameters.
//   - The MetadataServer must be running; the function uses Addr()
//     to build public URLs for each TTL file.
//   - If any file fails to ingest—or if directory traversal itself
//     fails—the function returns a non‑nil error.  The ingestion stops
//     at the first fatal directory walk error but continues on HTTP
//     errors.
func (ms *MetadataServer) PostFiles(gatewayURL, protocol string) error {
	gatewayURL = strings.TrimSuffix(gatewayURL, "/ui")
	postURL, err := url.Parse(gatewayURL)
	if err != nil {
		return fmt.Errorf("error parsing url '%s': %w", gatewayURL, err)
	}
	postURL = postURL.JoinPath("/populate")

	display.Done("Deployed metadata server with mounted dir: %s", ms.dir)
	display.Step("Starting the ingestion of the '*.ttl' files")

	ingestionError := false
	err = filepath.WalkDir(ms.dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			display.Error("Error while walking directory: %v", walkErr)
			ingestionError = true
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".ttl") {
			return nil
		}

		display.Step("Ingesting file: %s", d.Name())
		relPath, err := filepath.Rel(ms.dir, path)
		if err != nil {
			display.Error("Error getting relative path: %v", err)
			ingestionError = true
			return nil
		}

		q := postURL.Query()
		// TODO: remove securityCode once it's removed from the ingestor
		q.Set("securityCode", "changeme")
		q.Set("type", "single")
		q.Set("model", "EPOS-DCAT-AP-V1")
		q.Set("mapping", "EDM-TO-DCAT-AP")

		fileURL, err := url.JoinPath(protocol+"://", ms.Addr(), relPath)
		if err != nil {
			display.Error("Error while building URL for file '%s': %v", d.Name(), err)
			ingestionError = true
			return nil
		}

		q.Set("path", fileURL)
		postURL.RawQuery = q.Encode()

		r, err := http.NewRequest("POST", postURL.String(), nil)
		if err != nil {
			display.Error("Error building request for file '%s': %v", d.Name(), err)
			ingestionError = true
			return nil
		}

		r.Header.Add("accept", "*/*")
		res, err := http.DefaultClient.Do(r)
		if err != nil {
			display.Error("Error ingesting file '%s' in database: %v", d.Name(), err)
			ingestionError = true
			return nil
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			display.Error("Error reading response body for file '%s': %v", d.Name(), err)
			ingestionError = true
			return nil
		}

		if res.StatusCode != http.StatusOK {
			display.Error("Error ingesting file '%s' in database: received status code %d. Body of response: %s", d.Name(), res.StatusCode, string(body))
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

	display.Done("Ingestion of *.ttl files from dir '%s' finished successfully", ms.dir)
	return nil
}
