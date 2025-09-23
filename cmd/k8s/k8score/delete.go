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
	for _, podName := range opts.Name {
		eg.Go(func() error {
			display.Step("Deleting environment: %s", podName)

			env, err := db.GetKubernetesByName(podName)
			if err != nil {
				return fmt.Errorf("error getting kubernetes environment from db called '%s': %w", podName, err)
			}

			display.Step("Deleting namespace")

			if err := deleteNamespace(podName, env.Context); err != nil {
				return fmt.Errorf("error deleting namespace %s, %w", podName, err)
			}

			display.Done("Deleted namespace %s", podName)

			if err := common.RemoveEnvDir(env.Directory); err != nil {
				return fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
			}

			err = db.DeleteKubernetes(podName)
			if err != nil {
				return fmt.Errorf("failed to delete kubernetes %s (dir: %s) in db: %w", podName, env.Directory, err)
			}

			display.Done("Deleted environment: %s", podName)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (d *DeleteOpts) Validate() error {
	for _, podName := range d.Name {
		if err := validate.EnvironmentExistsK8s(podName); err != nil {
			return fmt.Errorf("invalid name for environment: %w", err)
		}
	}
	return nil
}
