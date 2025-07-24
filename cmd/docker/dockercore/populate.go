package dockercore

import (
	"fmt"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"
)

func Populate(name string, ttlDirs []string) (*db.Docker, error) {
	display.Step("Populating environment %s with %d directories", name, len(ttlDirs))

	docker, err := db.GetDockerByName(name)
	if err != nil {
		return nil, fmt.Errorf("error getting docker environment from db called '%s': %w", name, err)
	}

	for i, ttlDir := range ttlDirs {
		ttlDir, err = filepath.Abs(ttlDir)
		if err != nil {
			return nil, fmt.Errorf("error finding absolute path for given metadata path '%s': %w", ttlDir, err)
		}

		display.Step("Starting metadata server for directory %d of %d: %s", i+1, len(ttlDirs), ttlDir)
		metadataServer, err := common.NewMetadataServer(ttlDir)
		if err != nil {
			return nil, fmt.Errorf("creating metadata server for dir %q: %w", ttlDir, err)
		}

		if err = metadataServer.Start(); err != nil {
			return nil, fmt.Errorf("error starting metadata server: %w", err)
		}

		// Make sure the metadata server is stopped and URLs are printed last on success.
		defer func(env string) {
			display.Step("Stopping metadata server for directory: %s", ttlDir)
			if err := metadataServer.Stop(); err != nil {
				display.Error("Error while removing metadata server deployment: %v. You might have to remove it manually.", err)
			} else {
				display.Done("Metadata server stopped successfully")
			}
		}(name)

		err = metadataServer.PostFiles(docker.ApiUrl, "http")
		if err != nil {
			return nil, fmt.Errorf("error populating environment: %w", err)
		}
	}

	display.Done("Finished populating environment with ttl files from %d directories", len(ttlDirs))
	return docker, err
}
