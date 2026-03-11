package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
)

// PopulateOpts defines inputs for Populate.
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

// Populate ingests example or user-provided TTL data into an existing K8s environment.
func Populate(opts PopulateOpts) (*Env, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters for populate command: %w", err)
	}

	display.Step("Populating environment %s with %d directories", opts.Name, len(opts.TTLDirs))

	env, err := GetEnv(opts.Name, opts.Context)
	if err != nil {
		return nil, fmt.Errorf("error getting environment: %w", err)
	}

	display.Debug("loaded environment: %s", env.Name)

	port, err := common.FindFreePort()
	if err != nil {
		return nil, fmt.Errorf("error getting free port: %w", err)
	}

	display.Debug("selected local port for port-forward: %d", port)

	// start a port forward locally to the ingestor service and use that to do the populate posts
	err = ForwardAndRun(opts.Name, "ingestor-service", port, 8080, opts.Context, func(host string, port int) error {
		display.Step("Starting port-forward to ingestor-service pod")
		display.Debug("port-forward ready on %s:%d", host, port)

		url := fmt.Sprintf("http://%s:%d/api/ingestor-service/v1/", host, port)

		if opts.PopulateExamples {
			display.Debug("populating bundled examples through port-forward")

			successfulExamples, err := common.PopulateExample(url, opts.Parallel)
			if err != nil {
				return fmt.Errorf("error populating environment with examples through port-forward: %w", err)
			}

			display.Debug("populated example files: %d", len(successfulExamples))
		}

		for _, p := range opts.TTLDirs {
			display.Debug("processing metadata path: %s", p)

			absPath, err := filepath.Abs(p)
			if err != nil {
				return fmt.Errorf("error finding absolute path for given metadata path '%s': %w", p, err)
			}

			display.Debug("populating metadata from absolute path: %s", absPath)

			successfulFiles, err := common.PopulateEnv(absPath, url, opts.Parallel)
			if err != nil {
				return fmt.Errorf("error populating environment through port-forward: %w", err)
			}

			display.Debug("populated metadata files from path %s: %d", absPath, len(successfulFiles))
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error populating environment: %w", err)
	}

	display.Done("Finished populating environment with ttl files from %d directories", len(opts.TTLDirs))

	return env, nil
}

// Validate checks PopulateOpts and verifies paths, context, and environment state.
func (p *PopulateOpts) Validate() error {
	display.Debug("name: %s", p.Name)
	display.Debug("context: %s", p.Context)
	display.Debug("ttlDirs: %+v", p.TTLDirs)
	display.Debug("parallel: %d", p.Parallel)
	display.Debug("populateExamples: %v", p.PopulateExamples)

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
