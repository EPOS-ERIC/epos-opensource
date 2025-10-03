package common

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/epos-eu/epos-opensource/display"
	"golang.org/x/sync/errgroup"
)

// PopulateEnv ingests TTL (Turtle) files into an environment by posting them to the gatewey endpoint.
// It accepts either a single file or a directory path. When given a directory, it recursively walks through
// all subdirectories and ingests all *.ttl files found, processing them in parallel according to the
// specified concurrency limit.
//
// Parameters:
//   - ttlPath: Path to a TTL file or directory containing TTL files
//   - gatewayURL: Base URL of the EPOS gateway (e.g., "https://gateway.example.com")
//   - parallel: Maximum number of concurrent file ingestions (use 1 for sequential processing)
//
// Returns an error if any file fails to ingest or if the path is invalid.
func PopulateEnv(ttlPath, gatewayURL string, parallel int) error {
	if parallel == 0 {
		return fmt.Errorf("invalid parallel value: %d", parallel)
	}

	gatewayURL = strings.TrimSuffix(gatewayURL, "/ui")
	postURL, err := url.Parse(gatewayURL)
	if err != nil {
		return fmt.Errorf("invalid gateway URL '%s': %w", gatewayURL, err)
	}
	postURL = postURL.JoinPath("/populate")

	absPath, err := filepath.Abs(ttlPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	fi, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("cannot access path '%s': %w", ttlPath, err)
	}

	if fi.IsDir() {
		display.Step("Starting ingestion of *.ttl files from directory '%s'", ttlPath)

		walkError := false
		var eg errgroup.Group
		eg.SetLimit(parallel)

		err = filepath.WalkDir(ttlPath, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				display.Error("Failed to access path during directory traversal: %v", walkErr)
				walkError = true
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".ttl") {
				return nil
			}

			eg.Go(func() error {
				display.Step("Ingesting file: %s", d.Name())
				if err := postFile(path, *postURL); err != nil {
					display.Error("Failed to ingest '%s': %v", d.Name(), err)
					return err
				}
				return nil
			})
			return nil
		})
		if err != nil {
			return fmt.Errorf("directory traversal failed for '%s': %w", ttlPath, err)
		}
		if walkError {
			return fmt.Errorf("encountered errors while traversing directory '%s'", ttlPath)
		}
		if err := eg.Wait(); err != nil {
			return fmt.Errorf("one or more files failed to ingest in directory '%s': %w", ttlPath, err)
		}

		display.Done("Successfully ingested all *.ttl files from directory '%s'", ttlPath)
	} else {
		display.Step("Ingesting single file: %s", filepath.Base(ttlPath))
		err := postFile(absPath, *postURL)
		if err != nil {
			return fmt.Errorf("failed to ingest file '%s': %w", filepath.Base(ttlPath), err)
		}
		display.Done("Successfully ingested '%s'", filepath.Base(ttlPath))
	}

	return nil
}

var client = http.Client{
	Timeout: 60 * time.Second,
}

func postFile(path string, url url.URL) error {
	q := url.Query()
	// TODO: remove securityCode once it's removed from the ingestor
	q.Set("securityCode", "changeme")
	q.Set("type", "single")
	q.Set("model", "EPOS-DCAT-AP-V1")
	q.Set("mapping", "EDM-TO-DCAT-AP")
	url.RawQuery = q.Encode()

	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file '%s': %w", filepath.Base(path), err)
	}

	r, err := http.NewRequest("POST", url.String(), bytes.NewReader(file))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for '%s': %w", filepath.Base(path), err)
	}

	r.Header.Add("accept", "*/*")
	r.Header.Add("Content-Type", "text/turtle")

	res, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("HTTP request failed for '%s': %w", filepath.Base(path), err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body for '%s': %w", filepath.Base(path), err)
		}
		return fmt.Errorf("server rejected '%s' with status %d: %s", filepath.Base(path), res.StatusCode, string(body))
	}

	return nil
}
