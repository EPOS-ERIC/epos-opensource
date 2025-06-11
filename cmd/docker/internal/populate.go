package internal

import (
	"epos-cli/common"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

func Populate(path, name, ttlDir string) error {
	common.PrintStep("Populating environment: %s", name)
	dir, err := GetEnvDir(path, name)
	if err != nil {
		return fmt.Errorf("failed to get environment directory: %w", err)
	}
	common.PrintDone("Environment found in dir: %s", dir)

	common.PrintStep("Deploying metadata-cache")
	port, err := deployMetadataCache(ttlDir, name)
	if err != nil {
		common.PrintError("Failed to deploy metadata-cache: %v", err)
		return fmt.Errorf("failed to deploy metadata-cache: %w", err)
	}

	var success bool
	// make sure that the metadata-cache gets removed and URLs are printed last on success
	defer func(name, dir string) {
		common.PrintStep("Removing metadata-cache: %s-metadata-cache", name)
		err := deleteMetadataCache(name)
		if err != nil {
			common.PrintError("Error while removing deployed metadata cache: %v. You might have to remove it manually.", err)
		} else {
			common.PrintDone("Metadata-cache removed successfully")
		}

		if success {
			_ = common.PrintUrls(dir)
		}
	}(name, dir)

	postURL, err := getApiURL(dir)
	if err != nil {
		return fmt.Errorf("error getting api URL for env %s: %w", name, err)
	}
	postURL = postURL.JoinPath("/populate")

	common.PrintDone("Deployed metadata-cache with mounted dir: %s", ttlDir)
	common.PrintStep("Starting the ingestion of the '*.ttl' files")

	localIP, err := common.GetLocalIP()
	if err != nil {
		return fmt.Errorf("error getting local IP address: %w", err)
	}

	ingestionError := false
	err = filepath.WalkDir(ttlDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			common.PrintError("Error while walking directory: %v", err)
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

		fileURL, err := url.JoinPath("http://"+localIP+":"+strconv.Itoa(port), relPath)
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

		if res.StatusCode != 200 {
			common.PrintError("Error ingesting file '%s' in database: received status code %d. Body of response: %s", d.Name(), res.StatusCode, string(body))
			ingestionError = true
			return nil
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to ingest metadata in directory %s: %w", ttlDir, err)
	}

	if ingestionError {
		return fmt.Errorf("failed to ingest metadata in directory %s", ttlDir)
	}

	success = true
	return nil
}
