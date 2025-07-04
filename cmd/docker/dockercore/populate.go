package dockercore

import (
	"fmt"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
)

func Populate(name, ttlDir string) (*db.Docker, error) {
	common.PrintStep("Populating environment: %s", name)
	docker, err := db.GetDockerByName(name)
	if err != nil {
		return nil, fmt.Errorf("error getting docker environment from db called '%s': %w", name, err)
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

	err = metadataServer.PostFiles(docker.ApiUrl, "http")
	if err != nil {
		return nil, fmt.Errorf("error populating environment: %w", err)
	}

	return docker, err
}
