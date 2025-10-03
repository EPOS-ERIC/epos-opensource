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
	"github.com/epos-eu/epos-opensource/validate"
)

type PopulateOpts struct {
	// Required. name of the environment
	Name string
	// Required. list of directories or ttl files to populate the environment with
	TTLDirs []string
	// Optional. number of parallel uploads to do to the default is 1
	Parallel int
}

func Populate(opts PopulateOpts) (*sqlc.Kubernetes, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters for populate command: %w", err)
	}
	display.Step("Populating environment %s with %d directories", opts.Name, len(opts.TTLDirs))

	kube, err := db.GetKubernetesByName(opts.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes environment from db called '%s': %w", opts.Name, err)
	}

	for _, path := range opts.TTLDirs {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("error finding absolute path for given metadata path '%s': %w", path, err)
		}

		var metadataServer *metadataserver.MetadataServer

		metadataServer, err = metadataserver.New(path, opts.Parallel)
		if err != nil {
			return nil, fmt.Errorf("creating metadata server for file %q in directory none: %w", path, err)
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
		}(opts.Name)

		display.Step("Starting port-forward to ingestor-service pod")

		port, err := common.FindFreePort()
		if err != nil {
			return nil, fmt.Errorf("error getting free port: %w", err)
		}

		// start a port forward locally to the ingestor service and use that to do the populate posts
		err = ForwardAndRun(opts.Name, "ingestor-service", port, 8080, kube.Context, func(host string, port int) error {
			url := fmt.Sprintf("http://%s:%d/api/ingestor-service/v1/", host, port)

			err = metadataServer.PostFiles(url, kube.Protocol)
			if err != nil {
				return fmt.Errorf("error populating environment: %w", err)
			}

			return nil
		})
		if err != nil {
			display.Warn("error populating environment through port-forward, trying with direct URL: %s. error: %v", kube.ApiUrl, err)

			err = metadataServer.PostFiles(kube.ApiUrl, kube.Protocol)
			if err != nil {
				return nil, fmt.Errorf("error populating environment: %w", err)
			}
		}
	}

	display.Done("Finished populating environment with ttl files from %d directories", len(opts.TTLDirs))
	return kube, nil
}

func (p *PopulateOpts) Validate() error {
	if p.Parallel < 1 && p.Parallel > 20 {
		return fmt.Errorf("parallel uploads must be between 1 and 20")
	}

	if err := validate.EnvironmentExistsK8s(p.Name); err != nil {
		return fmt.Errorf("error validating environment name, no environment with '%s' exists: %w", p.Name, err)
	}

	for _, item := range p.TTLDirs {
		info, err := os.Stat(item)
		if err != nil {
			return fmt.Errorf("error stating path %q: %w", item, err)
		}
		if !info.IsDir() {
			if filepath.Ext(item) != ".ttl" {
				return fmt.Errorf("file %s is not a .ttl file", item)
			}
		}
	}

	return nil
}
