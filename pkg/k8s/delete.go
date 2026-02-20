package k8s

import (
	"fmt"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"golang.org/x/sync/errgroup"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// DeleteOpts TODO: add wait flag instead of by default? have it in the cli config?
type DeleteOpts struct {
	// Required. name of the environments to delete
	Name []string
	// Optional. Kubernetes context to use; defaults to the current kubectl context when unset.
	Context string
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
			display.Debug("building helm settings for env: %s", envName)

			settings := cli.New()
			settings.KubeContext = opts.Context
			settings.SetNamespace(envName)
			actionConfig := &action.Configuration{}

			if err := actionConfig.Init(
				settings.RESTClientGetter(),
				envName,
				"secret",
				func(format string, v ...any) { display.Info(format, v...) },
			); err != nil {
				return fmt.Errorf("failed to init helm action config for '%s': %w", envName, err)
			}

			display.Step("Uninstalling Helm release: %s", envName)
			display.Debug("running helm uninstall for env: %s", envName)

			uninstall := action.NewUninstall(actionConfig)
			uninstall.Wait = true
			uninstall.IgnoreNotFound = true
			uninstall.Timeout = 300 * time.Second // TODO set this globally for the package? have it in the config with a default?

			if _, err := uninstall.Run(envName); err != nil {
				return fmt.Errorf("failed to uninstall helm release '%s': %w", envName, err)
			}

			display.Done("Uninstalled Helm release: %s", envName)

			display.Step("Deleting namespace: %s", envName)
			display.Debug("running kubectl delete namespace for env: %s", envName)

			// helm does not delete the namespace, only its resources so we have to do it ourselves
			if err := deleteNamespace(envName, opts.Context); err != nil {
				return fmt.Errorf("failed to delete namespace %s: %w", envName, err)
			}

			display.Done("Deleted namespace: %s", envName)

			// TODO clean up ingested files

			display.Done("Deleted environment: %s", envName)

			return nil
		})
	}
	return eg.Wait()
}

func (d *DeleteOpts) Validate() error {
	display.Debug("names: %+v", d.Name)
	display.Debug("context: %s", d.Context)

	if d.Context == "" {
		context, err := common.GetCurrentKubeContext()
		if err != nil {
			return fmt.Errorf("failed to get current kubectl context: %w", err)
		}

		d.Context = context

		display.Debug("using current kubectl context: %s", d.Context)
	} else if err := EnsureContextExists(d.Context); err != nil {
		return fmt.Errorf("K8s context %q is not an available context: %w", d.Context, err)
	}

	for _, envName := range d.Name {
		if err := EnsureEnvironmentExists(envName, d.Context); err != nil {
			return fmt.Errorf("no environment with the name '%s' exists: %w", envName, err)
		}
	}

	return nil
}
