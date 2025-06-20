package internal

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/epos-eu/epos-opensource/common"
)

func Populate(customPath, name, ttlDir string) (portalURL, gatewayURL string, err error) {
	common.PrintStep("Populating environment: %s", name)

	dir, err := common.GetEnvDir(customPath, name, pathPrefix)
	if err != nil {
		return "", "", fmt.Errorf("failed to get environment directory: %w", err)
	}
	common.PrintDone("Environment found in dir: %s", dir)

	ttlDir, err = filepath.Abs(ttlDir)
	if err != nil {
		return "", "", fmt.Errorf("error finding absolute path for given metadata path '%s': %w", ttlDir, err)
	}

	common.PrintStep("Starting metadata server")
	metadataServer, err := common.NewMetadataServer(ttlDir)
	if err != nil {
		return "", "", fmt.Errorf("creating metadata server for dir %q: %w", ttlDir, err)
	}

	if err = metadataServer.Start(); err != nil {
		return "", "", fmt.Errorf("starting metadata server: %w", err)
	}

	// Make sure the metadata server is stopped and URLs are printed last on success.
	defer func(env string) {
		common.PrintStep("Stopping metadata server: %s-metadata-server", env)
		if err := metadataServer.Stop(); err != nil {
			common.PrintError("Error while removing metadata server deployment: %v. You might have to remove it manually.", err)
		} else {
			common.PrintDone("Metadata server removed successfully")
		}
	}(name)

	portalURL, gatewayURL, err = buildEnvURLs(dir)
	if err != nil {
		return "", "", fmt.Errorf("error building env urls for environment '%s': %w", dir, err)
	}

	postURL, err := url.Parse(gatewayURL)
	if err != nil {
		return "", "", fmt.Errorf("error parsing url '%s': %w", gatewayURL, err)
	}
	postURL = postURL.JoinPath("/populate")

	common.PrintDone("Deployed metadata server with mounted dir: %s", ttlDir)
	common.PrintStep("Starting the ingestion of the '*.ttl' files")

	ingestionError := false
	err = filepath.WalkDir(ttlDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			common.PrintError("Error while walking directory: %v", walkErr)
			ingestionError = true
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".ttl") {
			return nil
		}

		common.PrintStep("Ingesting file: %s", d.Name())
		relPath, err := filepath.Rel(ttlDir, path)
		if err != nil {
			common.PrintError("Error getting relative path: %v", err)
			ingestionError = true
			return nil
		}

		q := postURL.Query()
		// TODO: remove securityCode once it's removed from the ingestor
		q.Set("securityCode", "changeme")
		q.Set("type", "single")
		q.Set("model", "EPOS-DCAT-AP-V1")
		q.Set("mapping", "EDM-TO-DCAT-AP")

		fileURL, err := url.JoinPath("http://", metadataServer.Addr(), relPath)
		if err != nil {
			common.PrintError("Error while building URL for file '%s': %v", d.Name(), err)
			ingestionError = true
			return nil
		}

		q.Set("path", fileURL)
		postURL.RawQuery = q.Encode()

		r, err := http.NewRequest("POST", postURL.String(), nil)
		if err != nil {
			common.PrintError("Error building request for file '%s': %v", d.Name(), err)
			ingestionError = true
			return nil
		}

		r.Header.Add("accept", "*/*")
		res, err := http.DefaultClient.Do(r)
		if err != nil {
			common.PrintError("Error ingesting file '%s' in database: %v", d.Name(), err)
			ingestionError = true
			return nil
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			common.PrintError("Error reading response body for file '%s': %v", d.Name(), err)
			ingestionError = true
			return nil
		}

		if res.StatusCode != http.StatusOK {
			common.PrintError("Error ingesting file '%s' in database: received status code %d. Body of response: %s", d.Name(), res.StatusCode, string(body))
			ingestionError = true
			return nil
		}

		return nil
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to ingest metadata in directory %s: %w", ttlDir, err)
	}

	if ingestionError {
		return "", "", fmt.Errorf("failed to ingest metadata in directory %s", ttlDir)
	}

	gatewayURL, err = url.JoinPath(gatewayURL, "ui/")
	return portalURL, gatewayURL, err
}
