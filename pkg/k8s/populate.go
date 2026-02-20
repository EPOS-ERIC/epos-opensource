package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/db"
)

type PopulateOpts struct {
	// Required. name of the environment
	Name string
	// Optional. Kubernetes context to use; defaults to the current kubectl context when unset.
	Context string
	// Required. list of directories or ttl files to populate the environment with
	TTLDirs []string
	// Optional. number of parallel uploads to do to the default is 1
	Parallel int
	// Optional. weather to populate the examples or not
	PopulateExamples bool
}

func Populate(opts PopulateOpts) (*Env, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters for populate command: %w", err)
	}

	display.Step("Populating environment %s with %d directories", opts.Name, len(opts.TTLDirs))

	env, err := GetEnv(opts.Name, opts.Context)
	if err != nil {
		return nil, fmt.Errorf("error getting environment: %w", err)
	}

	port, err := common.FindFreePort()
	if err != nil {
		return nil, fmt.Errorf("error getting free port: %w", err)
	}

	URLs, err := env.BuildEnvURLs()
	if err != nil {
		return nil, fmt.Errorf("error building environment URLs: %w", err)
	}

	// start a port forward locally to the ingestor service and use that to do the populate posts
	err = ForwardAndRun(opts.Name, "ingestor-service", port, 8080, opts.Context, func(host string, port int) error {
		display.Step("Starting port-forward to ingestor-service pod")

		url := fmt.Sprintf("http://%s:%d/api/ingestor-service/v1/", host, port)

		var allSuccessfulFiles []string

		if opts.PopulateExamples {
			successfulExamples, err := common.PopulateExample(url, opts.Parallel)
			allSuccessfulFiles = append(allSuccessfulFiles, successfulExamples...)
			if err != nil {
				display.Warn("error populating environment with examples through port-forward: %v. Trying with direct URL: %s", err, URLs.APIURL)

				successfulExamples, err = common.PopulateExample(URLs.APIURL, opts.Parallel)
				allSuccessfulFiles = append(allSuccessfulFiles, successfulExamples...)
				if err != nil {
					return fmt.Errorf("error populating environment with examples with direct URL: %w", err)
				}
			}
		}

		for _, p := range opts.TTLDirs {
			absPath, err := filepath.Abs(p)
			if err != nil {
				return fmt.Errorf("error finding absolute path for given metadata path '%s': %w", p, err)
			}

			successfulFiles, err := common.PopulateEnv(absPath, url, opts.Parallel)
			allSuccessfulFiles = append(allSuccessfulFiles, successfulFiles...)
			if err != nil {
				display.Warn("error populating environment through port-forward: %v. Trying with direct URL: %s", err, URLs.APIURL)
				successfulFiles, err = common.PopulateEnv(absPath, URLs.APIURL, opts.Parallel)
				allSuccessfulFiles = append(allSuccessfulFiles, successfulFiles...)
				if err != nil {
					return fmt.Errorf("error populating environment with direct URL: %w", err)
				}
			}
		}

		// Insert ingested files into database
		for _, filePath := range allSuccessfulFiles {
			if err := db.InsertIngestedFile("k8s", opts.Name, filePath); err != nil {
				return fmt.Errorf("error inserting ingested file record: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error populating environment: %w", err)
	}

	display.Done("Finished populating environment with ttl files from %d directories", len(opts.TTLDirs))

	return env, nil
}

func (p *PopulateOpts) Validate() error {
	if p.Parallel < 1 || p.Parallel > 20 {
		return fmt.Errorf("parallel uploads must be between 1 and 20")
	}

	if p.Context == "" {
		context, err := common.GetCurrentKubeContext()
		if err != nil {
			return fmt.Errorf("failed to get current kubectl context: %w", err)
		}

		p.Context = context
	} else if err := EnsureContextExists(p.Context); err != nil {
		return fmt.Errorf("K8s context %q is not an available context: %w", p.Context, err)
	}

	if err := EnsureEnvironmentExists(p.Name, p.Context); err != nil {
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
