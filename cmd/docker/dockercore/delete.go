package dockercore

import (
	_ "embed"
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/epos-eu/epos-opensource/validate"
)

type DeleteOpts struct {
	// Required. name of the environment
	Name string
}

func Delete(opts DeleteOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid delete parameters: %w", err)
	}
	display.Step("Deleting environment: %s", opts.Name)

	env, err := db.GetDockerByName(opts.Name)
	if err != nil {
		return fmt.Errorf("error getting docker environment from db called '%s': %w", opts.Name, err)
	}

	display.Step("Stopping stack")

	if err := downStack(env.Directory, true); err != nil {
		return fmt.Errorf("docker compose down failed: %w", err)
	}

	display.Done("Stopped environment: %s", opts.Name)

	if err := common.RemoveEnvDir(env.Directory); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
	}

	err = db.DeleteDocker(opts.Name)
	if err != nil {
		return fmt.Errorf("failed to delete docker %s (dir: %s) in db: %w", opts.Name, env.Directory, err)
	}

	display.Done("Deleted environment: %s", opts.Name)

	return nil
}
func (d *DeleteOpts) Validate() error {
	if err := validate.EnviromentExistsDocker(d.Name); err != nil {
		return fmt.Errorf("no environment with the name '%s' exists: %w", d.Name, err)
	}
	return nil
}
