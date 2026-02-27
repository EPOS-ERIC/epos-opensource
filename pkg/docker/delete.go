package docker

import (
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/EPOS-ERIC/epos-opensource/db"
	"github.com/EPOS-ERIC/epos-opensource/display"
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
			display.Debug("loading docker environment: %s", envName)

			env, err := GetEnv(envName)
			if err != nil {
				return fmt.Errorf("error getting docker environment '%s': %w", envName, err)
			}

			display.Step("Stopping stack for environment: %s", envName)
			display.Debug("running docker compose down for environment: %s", envName)

			if err := downStack(&env.EnvConfig, true); err != nil {
				return fmt.Errorf("docker compose down failed for '%s': %w", envName, err)
			}

			display.Done("Stopped environment: %s", envName)

			display.Debug("deleting ingested file records for: %s", envName)

			if err := db.DeleteIngestedFilesByEnvironment(envName); err != nil {
				return fmt.Errorf("failed to delete ingested files for '%s': %w", envName, err)
			}

			display.Debug("deleting docker environment record for: %s", envName)

			if err := db.DeleteDocker(envName); err != nil {
				return fmt.Errorf("failed to delete docker '%s' in db: %w", envName, err)
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
