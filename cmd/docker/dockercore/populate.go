package dockercore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/epos-eu/epos-opensource/metadataserver"
	"github.com/epos-eu/epos-opensource/validate"
)

type PopulateOpts struct {
	// Required. paths to ttl files or directories containing ttl files to populate the environment
	TTLDirs []string
	// Required. name of the environment to populate
	Name string
	// Optional. number of parallel uploads to use. If not set the default will be 1. Max is 20
	Parallel int
}

func Populate(opts PopulateOpts) (*sqlc.Docker, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid populate parameters: %w", err)
	}
	display.Step("Populating environment %s with %d path(s)", opts.Name, len(opts.TTLDirs))
	docker, err := db.GetDockerByName(opts.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting docker environment from db called '%s': %w", opts.Name, err)
	}

	for _, p := range opts.TTLDirs {
		p, err = filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("error finding absolute path for given metadata path '%s': %w", p, err)
		}

		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("error stating path %q: %w", p, err)
		}

		var metadataServer *metadataserver.MetadataServer
		if !info.IsDir() {
			if filepath.Ext(p) != ".ttl" {
				return nil, fmt.Errorf("file %s is not a .ttl file", p)
			}
		}
		metadataServer, err = metadataserver.New(p, opts.Parallel)
		if err != nil {
			return nil, fmt.Errorf("creating metadata server for dir %q: %w", p, err)
		}

		if err = metadataServer.Start(); err != nil {
			return nil, fmt.Errorf("error starting metadata server: %w", err)
		}

		defer func(env string) {
			display.Step("Stopping metadata server for path: %s", p)
			if err := metadataServer.Stop(); err != nil {
				display.Error("Error while removing metadata server deployment: %v. You might have to remove it manually.", err)
			} else {
				display.Done("Metadata server stopped successfully")
			}
		}(opts.Name)

		err = metadataServer.PostFiles(docker.ApiUrl, "http")
		if err != nil {
			return nil, fmt.Errorf("error populating environment: %w", err)
		}
	}

	display.Done("Finished populating environment with ttl files from %d path(s)", len(opts.TTLDirs))
	return docker, err
}
func (p *PopulateOpts) Validate() error {
	if p.Parallel < 1 && p.Parallel > 20 {
		return fmt.Errorf("parallel uploads must be between 1 and 20")
	}
	if validate.EnvironmentExistsDocker(p.Name) != nil {
		return fmt.Errorf("no environment with name'%s' exists", p.Name)
	}

	return nil
}
