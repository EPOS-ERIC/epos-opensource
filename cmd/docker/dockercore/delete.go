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
	Name []string
}

func Delete(opts DeleteOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid delete parameters: %w", err)
	}
	num_enviroment := len(opts.Name)
	for i, d := range opts.Name {
		display.Step("Deleting environment: %s", d)

		env, err := db.GetDockerByName(d)
		if err != nil {
			return fmt.Errorf("error getting docker environment from db called '%s': %w", d, err)
		}
		display.Step("Stopping stack")

		display.Step("Deleting enviroment: %d of %d", i+1, num_enviroment)

		if err := downStack(env.Directory, true); err != nil {
			return fmt.Errorf("docker compose down failed: %w", err)
		}

		display.Done("Stopped environment: %s", d)

		if err := common.RemoveEnvDir(env.Directory); err != nil {
			return fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
		}

		err = db.DeleteDocker(d)
		if err != nil {
			return fmt.Errorf("failed to delete docker %s (dir: %s) in db: %w", d, env.Directory, err)
		}

		display.Done("Deleted environment: %s", d)
	}
	return nil
}

func (d *DeleteOpts) Validate() error {
	for _, env := range d.Name {
		if err := validate.EnvironmentExistsDocker(env); err != nil {
			return fmt.Errorf("no environment with the name '%s' exists: %w", env, err)
		}
	}

	return nil
}
