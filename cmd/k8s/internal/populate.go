package internal

import (
	"epos-opensource/common"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
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

	common.PrintStep("Deploying metadata-cache")
	err = deployMetadataCache(dir, ttlDir, name)
	if err != nil {
		common.PrintError("Failed to deploy metadata-cache: %v", err)
		return "", "", fmt.Errorf("failed to deploy metadata-cache: %w", err)
	}

	// make sure that the metadata-cache gets removed and URLs are printed last on success
	defer func(name string) {
		common.PrintStep("Removing metadata-cache: %s-metadata-cache", name)
		err := deleteMetadataCache(dir, name)
		if err != nil {
			common.PrintError("Error while removing deployed metadata cache: %v. You might have to remove it manually.", err)
		} else {
			common.PrintDone("Metadata-cache removed successfully")
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

	common.PrintDone("Deployed metadata-cache with mounted dir: %s", ttlDir)
	common.PrintStep("Starting the ingestion of the '*.ttl' files")

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

		fileURL, err := url.JoinPath("http://metadata-cache:80", relPath)
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
		return "", "", fmt.Errorf("failed to ingest metadata in directory %s: %w", ttlDir, err)
	}

	if ingestionError {
		return "", "", fmt.Errorf("failed to ingest metadata in directory %s", ttlDir)
	}

	gatewayURL, err = url.JoinPath(gatewayURL, "ui/")
	return portalURL, gatewayURL, err
}

// deployMetadataCache deploys an nginx docker container running a file server exposing a volume
func deployMetadataCache(dir, ttlDir, envName string) error {
	metadataCachePath := path.Join(dir, "metadata-cache.yaml")
	os.Setenv("TTL_FILES_DIR_PATH", ttlDir)
	os.Setenv("NAMESPACE", envName)

	expandedManifest := os.ExpandEnv(metadataCacheManifest)
	err := os.WriteFile(metadataCachePath, []byte(expandedManifest), 0777)
	if err != nil {
		return fmt.Errorf("error writing metadata-cache manifest to dir '%s': %w", metadataCachePath, err)
	}

	err = runKubectl(dir, false, "apply", "-f", "metadata-cache.yaml")
	if err != nil {
		return fmt.Errorf("error applying metadata-cache deployment: %w", err)
	}

	err = waitDeployments(dir, envName, []string{"metadata-cache"})
	if err != nil {
		return fmt.Errorf("error while waiting for metadata-cache deployment to start: %w", err)
	}

	return nil
}

// deleteMetadataCache removes a deployment of a metadata cache container
func deleteMetadataCache(dir, envName string) error {
	// Delete the deployment
	err := runKubectl(dir, false, "delete", "-f", "metadata-cache.yaml")
	if err != nil {
		return fmt.Errorf("error deleting metadata-cache deployment: %w", err)
	}

	// Wait for it to be completely gone
	err = runKubectl(dir, false, "wait", "--for=delete", "deployment/metadata-cache",
		"--timeout=60s", "-n", envName)
	if err != nil {
		return fmt.Errorf("error waiting for metadata-cache deployment to be deleted: %w", err)
	}

	// delete the manifest
	metadataCachePath := path.Join(dir, "metadata-cache.yaml")

	err = os.Remove(metadataCachePath)
	if err != nil {
		return fmt.Errorf("error removing metadata-cache.yaml: %w", err)
	}

	return nil
}
