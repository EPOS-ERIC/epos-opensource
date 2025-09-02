package k8score

import (
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
		return fmt.Errorf("invalid parameters for delete command: %w", err)
	}
	display.Step("Deleting environment: %s", opts.Name)

	env, err := db.GetKubernetesByName(opts.Name)
	if err != nil {
		return fmt.Errorf("error getting kubernetes environment from db called '%s': %w", opts.Name, err)
	}

	display.Step("Deleting namespace")

	if err := deleteNamespace(opts.Name, env.Context); err != nil {
		return fmt.Errorf("error deleting namespace %s, %w", opts.Name, err)
	}

	display.Done("Deleted namespace %s", opts.Name)

	if err := common.RemoveEnvDir(env.Directory); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", env.Directory, err)
	}

	err = db.DeleteKubernetes(opts.Name)
	if err != nil {
		return fmt.Errorf("failed to delete kubernetes %s (dir: %s) in db: %w", opts.Name, env.Directory, err)
	}

	display.Done("Deleted environment: %s", opts.Name)

	return nil
}
func (d *DeleteOpts) Validate() error {
	if err := validate.EnviromentNotExistK8s(d.Name); err != nil {
		return fmt.Errorf("invalid name for environment: %w", err)
	}
	return nil
}
