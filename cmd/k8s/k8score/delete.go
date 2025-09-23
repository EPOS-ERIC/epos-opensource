package k8score

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/epos-eu/epos-opensource/validate"
	"golang.org/x/sync/errgroup"
)

type DeleteOpts struct {
	// Required. name of the environment
	Name []string
}

func Delete(opts DeleteOpts) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid parameters for delete command: %w", err)
	}
	var eg errgroup.Group
	eg.SetLimit(20)
	for _, envName := range opts.Name {
		eg.Go(func() error {
			display.Step("Deleting environment: %s", envName)

			env, err := db.GetKubernetesByName(envName)
			if err != nil {
				return fmt.Errorf("error getting kubernetes environment from db called '%s': %w", envName, err)
			}

			display.Step("Deleting namespace")

			if err := deleteNamespace(envName, env.Context); err != nil {
				return fmt.Errorf("error deleting namespace %s, %w", envName, err)
			}

			display.Done("Deleted namespace %s", envName)

			if err := common.RemoveEnvDir(env.Directory); err != nil {
				return fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
			}

			err = db.DeleteKubernetes(envName)
			if err != nil {
				return fmt.Errorf("failed to delete kubernetes %s (dir: %s) in db: %w", envName, env.Directory, err)
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
	for _, envName := range d.Name {
		if err := validate.EnvironmentExistsK8s(envName); err != nil {
			return fmt.Errorf("invalid name for environment: %w", err)
		}
	}
	return nil
}
