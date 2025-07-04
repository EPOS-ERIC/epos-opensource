package k8score

import (
	"fmt"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
)

func Populate(name, ttlDir string) (*db.Kubernetes, error) {
	common.PrintStep("Populating environment: %s", name)

	kube, err := db.GetKubernetesByName(name)
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes environment from db called '%s': %w", name, err)
	}

	ttlDir, err = filepath.Abs(ttlDir)
	if err != nil {
		return nil, fmt.Errorf("error finding absolute path for given metadata path '%s': %w", ttlDir, err)
	}

	common.PrintStep("Starting metadata server")
	metadataServer, err := common.NewMetadataServer(ttlDir)
	if err != nil {
		return nil, fmt.Errorf("creating metadata server for dir %q: %w", ttlDir, err)
	}

	if err = metadataServer.Start(); err != nil {
		return nil, fmt.Errorf("starting metadata server: %w", err)
	}

	// Make sure the metadata server is stopped and URLs are printed last on success.
	defer func(env string) {
		common.PrintStep("Stopping metadata server")
		if err := metadataServer.Stop(); err != nil {
			common.PrintError("Error while removing metadata server deployment: %v. You might have to remove it manually.", err)
		} else {
			common.PrintDone("Metadata server stopped successfully")
		}
	}(name)

	common.PrintStep("Starting port-forward to ingestor-service pod")
	port, err := common.FreePort()
	if err != nil {
		return nil, fmt.Errorf("error getting free port: %w", err)
	}
	// start a port forward locally to the ingestor service and use that to do the populate posts
	err = ForwardAndRun(name, "ingestor-service", port, 8080, kube.Context, func(host string, port int) error {
		common.PrintDone("Port forward started successfully")
		// here we use http because we are accessing the apis in the pod itself, through the port forward
		url := fmt.Sprintf("http://%s:%d/api/ingestor-service/v1/", host, port)
		err = metadataServer.PostFiles(url, kube.Protocol)
		if err != nil {
			return fmt.Errorf("error populating environment: %w", err)
		}
		return nil
	})
	if err != nil {
		common.PrintWarn("error populating environment through port-forward, trying with direct IP. error: %v", err)
		err = metadataServer.PostFiles(kube.ApiUrl, kube.Protocol)
		if err != nil {
			return nil, fmt.Errorf("error populating environment: %w", err)
		}
	}

	return kube, nil
}
