package dockercore

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

	if opts.PopulateExamples {
		err := common.PopulateExample(docker.ApiUrl, opts.Parallel)
		if err != nil {
			return nil, fmt.Errorf("error populating environment with examples: %w", err)
		}
	}

	for _, p := range opts.TTLDirs {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("error finding absolute path for given metadata path '%s': %w", p, err)
		}

		if err := common.PopulateEnv(absPath, docker.ApiUrl, opts.Parallel); err != nil {
			return nil, fmt.Errorf("error populating environment: %w", err)
		}
	}

	display.Done("Finished populating environment with ttl files from %d path(s)", len(opts.TTLDirs))
	return docker, nil
}

func (p *PopulateOpts) Validate() error {
	if p.Parallel < 1 && p.Parallel > 20 {
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
