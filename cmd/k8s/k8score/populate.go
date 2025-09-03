package k8score

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/epos-eu/epos-opensource/metadataserver"
)

func Populate(name string, ttlDirs []string, parallel int) (*sqlc.Kubernetes, error) {
	display.Step("Populating environment %s with %d directories", name, len(ttlDirs))

	kube, err := db.GetKubernetesByName(name)
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes environment from db called '%s': %w", name, err)
	}

	for i, path := range ttlDirs {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("error finding absolute path for given metadata path '%s': %w", path, err)
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("error stating path %q: %w", path, err)
		}
		var metadataServer *metadataserver.MetadataServer

		if info.IsDir() {
			// Case 1: directory
			display.Step("Starting metadata server for directory %d of %d: %s", i+1, len(ttlDirs), path)
			metadataServer, err = metadataserver.New(path, parallel)
			if err != nil {
				return nil, fmt.Errorf("creating metadata server for dir %q: %w", path, err)
			}
		} else {
			// Case 2: file
			if filepath.Ext(path) != ".ttl" {
				return nil, fmt.Errorf("file %s is not a .ttl file", path)
			}

			metadataServer, err = metadataserver.New(path, parallel)
			if err != nil {
				return nil, fmt.Errorf("creating metadata server for file %q in directory none: %w", path, err)
			}

		}

		if err = metadataServer.Start(); err != nil {
			return nil, fmt.Errorf("starting metadata server: %w", err)
		}

		// Make sure the metadata server is stopped and URLs are printed last on success.
		defer func(env string) {
			display.Step("Stopping metadata server for directory: %s", path)
			if err := metadataServer.Stop(); err != nil {
				display.Error("Error while removing metadata server deployment: %v. You might have to remove it manually.", err)
			} else {
				display.Done("Metadata server stopped successfully")
			}
		}(name)

		display.Step("Starting port-forward to ingestor-service pod")
		port, err := common.FindFreePort()
		if err != nil {
			return nil, fmt.Errorf("error getting free port: %w", err)
		}
		// start a port forward locally to the ingestor service and use that to do the populate posts
		err = ForwardAndRun(name, "ingestor-service", port, 8080, kube.Context, func(host string, port int) error {
			display.Done("Port forward started successfully")
			// here we use http because we are accessing the apis in the pod itself, through the port forward
			url := fmt.Sprintf("http://%s:%d/api/ingestor-service/v1/", host, port)
			err = metadataServer.PostFiles(url, kube.Protocol)
			if err != nil {
				return fmt.Errorf("error populating environment: %w", err)
			}
			return nil
		})
		if err != nil {
			display.Warn("error populating environment through port-forward, trying with direct IP. error: %v", err)
			err = metadataServer.PostFiles(kube.ApiUrl, kube.Protocol)
			if err != nil {
				return nil, fmt.Errorf("error populating environment: %w", err)
			}
		}
	}

	display.Done("Finished populating environment with ttl files from %d directories", len(ttlDirs))
	return kube, nil
}
