package docker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/db"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/db/sqlc"
)

// PopulateOpts defines inputs for Populate.
type PopulateOpts struct {
	// Required. paths to ttl files or directories containing ttl files to populate the environment
	TTLDirs []string
	// Required. name of the environment to populate
	Name string
	// Optional. number of parallel uploads to use. If not set the default will be 1. Max is 20
	Parallel int
	// Optional. weather to populate the examples or not
	PopulateExamples bool
}

// Populate ingests example or user-provided TTL data into an existing Docker environment.
func Populate(opts PopulateOpts) (*sqlc.Docker, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid populate parameters: %w", err)
	}

	display.Step("Populating environment %s with %d path(s)", opts.Name, len(opts.TTLDirs))

	docker, err := db.GetDockerByName(opts.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting docker environment from db called '%s': %w", opts.Name, err)
	}

	display.Debug("loaded docker environment: %s (api: %s)", docker.Name, docker.ApiUrl)

	var allSuccessfulFiles []string

	if opts.PopulateExamples {
		display.Debug("populating bundled examples")

		successfulExamples, err := common.PopulateExample(docker.ApiUrl, opts.Parallel)
		allSuccessfulFiles = append(allSuccessfulFiles, successfulExamples...)
		if err != nil {
			return nil, fmt.Errorf("error populating environment with examples: %w", err)
		}

		display.Debug("populated example files: %d", len(successfulExamples))
	}

	for _, p := range opts.TTLDirs {
		display.Debug("processing metadata path: %s", p)

		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("error finding absolute path for given metadata path '%s': %w", p, err)
		}

		display.Debug("populating metadata from absolute path: %s", absPath)

		successfulFiles, err := common.PopulateEnv(absPath, docker.ApiUrl, opts.Parallel)
		allSuccessfulFiles = append(allSuccessfulFiles, successfulFiles...)
		if err != nil {
			return nil, fmt.Errorf("error populating environment: %w", err)
		}

		display.Debug("populated metadata files from path %s: %d", absPath, len(successfulFiles))
	}

	// Insert ingested files into database
	for _, filePath := range allSuccessfulFiles {
		display.Debug("recording ingested file: %s", filePath)

		if err := db.InsertIngestedFile(opts.Name, filePath); err != nil {
			return nil, fmt.Errorf("error inserting ingested file record: %w", err)
		}
	}

	display.Debug("recorded ingested files: %d", len(allSuccessfulFiles))
	display.Done("Finished populating environment with ttl files from %d path(s)", len(opts.TTLDirs))

	return docker, nil
}

// Validate checks PopulateOpts and verifies that input paths and environment state are valid.
func (p *PopulateOpts) Validate() error {
	display.Debug("name: %s", p.Name)
	display.Debug("ttlDirs: %+v", p.TTLDirs)
	display.Debug("parallel: %d", p.Parallel)
	display.Debug("populateExamples: %v", p.PopulateExamples)

	if p.Parallel < 1 || p.Parallel > 20 {
		return fmt.Errorf("parallel uploads must be between 1 and 20")
	}

	if err := EnsureEnvironmentExists(p.Name); err != nil {
		return fmt.Errorf("no environment with name '%s' exists: %w", p.Name, err)
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
