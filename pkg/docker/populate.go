package docker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/db"
	"github.com/EPOS-ERIC/epos-opensource/db/sqlc"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/validate"
)

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

func Populate(opts PopulateOpts) (*sqlc.Docker, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid populate parameters: %w", err)
	}

	display.Step("Populating environment %s with %d path(s)", opts.Name, len(opts.TTLDirs))

	docker, err := db.GetDockerByName(opts.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting docker environment from db called '%s': %w", opts.Name, err)
	}

	var allSuccessfulFiles []string

	if opts.PopulateExamples {
		successfulExamples, err := common.PopulateExample(docker.ApiUrl, opts.Parallel)
		allSuccessfulFiles = append(allSuccessfulFiles, successfulExamples...)
		if err != nil {
			return nil, fmt.Errorf("error populating environment with examples: %w", err)
		}
	}

	for _, p := range opts.TTLDirs {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("error finding absolute path for given metadata path '%s': %w", p, err)
		}

		successfulFiles, err := common.PopulateEnv(absPath, docker.ApiUrl, opts.Parallel)
		allSuccessfulFiles = append(allSuccessfulFiles, successfulFiles...)
		if err != nil {
			return nil, fmt.Errorf("error populating environment: %w", err)
		}
	}

	// Insert ingested files into database
	for _, filePath := range allSuccessfulFiles {
		if err := db.InsertIngestedFile("docker", opts.Name, filePath); err != nil {
			return nil, fmt.Errorf("error inserting ingested file record: %w", err)
		}
	}

	display.Done("Finished populating environment with ttl files from %d path(s)", len(opts.TTLDirs))
	return docker, nil
}

func (p *PopulateOpts) Validate() error {
	if p.Parallel < 1 || p.Parallel > 20 {
		return fmt.Errorf("parallel uploads must be between 1 and 20")
	}

	if validate.EnvironmentExistsDocker(p.Name) != nil {
		return fmt.Errorf("no environment with name'%s' exists", p.Name)
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
