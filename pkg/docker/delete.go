package docker

import (
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/db"
)

// DeleteOpts defines inputs for Delete.
type DeleteOpts struct {
	Name []string // names of environments
}

// Delete stops and removes one or more Docker environments and their tracked metadata.
func Delete(opts DeleteOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid delete parameters: %w", err)
	}

	var eg errgroup.Group
	eg.SetLimit(20)
	for _, envName := range opts.Name {
		eg.Go(func() error {
			display.Step("Deleting environment: %s", envName)
			display.Debug("loading docker environment from db: %s", envName)

			env, err := db.GetDockerByName(envName)
			if err != nil {
				return fmt.Errorf("error getting docker environment '%s': %w", envName, err)
			}

			display.Debug("loaded docker environment directory: %s", env.Directory)
			display.Step("Stopping stack for environment: %s", envName)
			display.Debug("running docker compose down for: %s", env.Directory)

			if err := downStack(env.Directory, true); err != nil {
				return fmt.Errorf("docker compose down failed for '%s': %w", envName, err)
			}

			display.Done("Stopped environment: %s", envName)
			display.Debug("removing environment directory: %s", env.Directory)

			if err := common.RemoveEnvDir(env.Directory); err != nil {
				return fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
			}

			display.Debug("deleting ingested file records for: %s", envName)

			if err := db.DeleteIngestedFilesByEnvironment("docker", envName); err != nil {
				return fmt.Errorf("failed to delete ingested files for '%s': %w", envName, err)
			}

			display.Debug("deleting docker environment record for: %s", envName)

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

// Validate checks DeleteOpts and ensures every requested environment exists.
func (d *DeleteOpts) Validate() error {
	display.Debug("names: %+v", d.Name)

	for _, env := range d.Name {
		if err := EnsureEnvironmentExists(env); err != nil {
			return fmt.Errorf("no environment with the name '%s' exists: %w", env, err)
		}
	}
	return nil
}
