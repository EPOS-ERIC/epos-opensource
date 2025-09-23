package dockercore

import (
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/epos-eu/epos-opensource/validate"
)

type DeleteOpts struct {
	Name []string // names of environments
}

func Delete(opts DeleteOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid delete parameters: %w", err)
	}

	var eg errgroup.Group
	eg.SetLimit(20)
	for _, envName := range opts.Name {
		eg.Go(func() error {
			display.Step("Deleting environment: %s", envName)

			env, err := db.GetDockerByName(envName)
			if err != nil {
				return fmt.Errorf("error getting docker environment '%s': %w", envName, err)
			}

			display.Step("Stopping stack for environment: %s", envName)
			if err := downStack(env.Directory, true); err != nil {
				return fmt.Errorf("docker compose down failed for '%s': %w", envName, err)
			}
			display.Done("Stopped environment: %s", envName)

			if err := common.RemoveEnvDir(env.Directory); err != nil {
				return fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
			}

			if err := db.DeleteDocker(envName); err != nil {
				return fmt.Errorf("failed to delete docker '%s' (dir: %s) in db: %w", envName, env.Directory, err)
			}

			display.Done("Deleted environment: %s", envName)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
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
