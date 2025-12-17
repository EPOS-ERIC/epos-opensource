package k8score

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/epos-eu/epos-opensource/validate"
)

type PopulateOpts struct {
	// Required. name of the environment
	Name string
	// Required. list of directories or ttl files to populate the environment with
	TTLDirs []string
	// Optional. number of parallel uploads to do to the default is 1
	Parallel int
	// Optional. weather to populate the examples or not
	PopulateExamples bool
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

	port, err := common.FindFreePort()
	if err != nil {
		return nil, fmt.Errorf("error getting free port: %w", err)
	}
	// start a port forward locally to the ingestor service and use that to do the populate posts
	err = ForwardAndRun(opts.Name, "ingestor-service", port, 8080, kube.Context, func(host string, port int) error {
		display.Step("Starting port-forward to ingestor-service pod")

		url := fmt.Sprintf("http://%s:%d/api/ingestor-service/v1/", host, port)

		var allSuccessfulFiles []string

		if opts.PopulateExamples {
			successfulExamples, err := common.PopulateExample(url, opts.Parallel)
			if err != nil {
				display.Warn("error populating environment with examples through port-forward: %v. Trying with direct URL: %s", err, kube.ApiUrl)
				successfulExamples, err = common.PopulateExample(kube.ApiUrl, opts.Parallel)
				if err != nil {
					return fmt.Errorf("error populating environment with examples with direct URL: %w", err)
				}
			}
			allSuccessfulFiles = append(allSuccessfulFiles, successfulExamples...)
		}

		for _, path := range opts.TTLDirs {
			path, err = filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("error finding absolute path for given metadata path '%s': %w", path, err)
			}

			successfulFiles, err := common.PopulateEnv(path, url, opts.Parallel)
			if err != nil {
				display.Warn("error populating environment through port-forward: %v. Trying with direct URL: %s", err, kube.ApiUrl)
				successfulFiles, err = common.PopulateEnv(path, kube.ApiUrl, opts.Parallel)
				if err != nil {
					return fmt.Errorf("error populating environment with direct URL: %w", err)
				}
			}
			allSuccessfulFiles = append(allSuccessfulFiles, successfulFiles...)
		}

		// Insert ingested files into database
		for _, filePath := range allSuccessfulFiles {
			if err := db.InsertIngestedFile("kubernetes", opts.Name, filePath); err != nil {
				return fmt.Errorf("error inserting ingested file record: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error populating environment: %w", err)
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
