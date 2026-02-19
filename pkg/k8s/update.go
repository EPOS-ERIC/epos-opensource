// Package k8s contains the internal functions used by the k8s cmd to manage the environments
package k8s

import (
	_ "embed"
	"fmt"
	"slices"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
	"github.com/EPOS-ERIC/epos-opensource/validate"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type UpdateOpts struct {
	// Required. name of the environment
	OldEnvName string
	// Optional. Kubernetes context to use; defaults to the current kubectl context when unset.
	Context string
	// Optional. do an uninstall then an install (to wipe volumes and stuff) maybe we should rename the flag?
	Force bool
	// Optional. reset the environment config to the embedded defaults maybe we should rename the flag?
	Reset bool
	// New configuration to apply. If nil, preserves existing config
	NewConfig *config.Config
}

// Update TODO: add docs
func Update(opts UpdateOpts) (*Env, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters for update command: %w", err)
	}

	if opts.NewConfig != nil && opts.NewConfig.Name == "" {
		opts.NewConfig.Name = opts.OldEnvName
	}

	if opts.NewConfig == nil {
		oldEnv, err := GetEnv(opts.OldEnvName, opts.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to get environment: %w", err)
		}

		opts.NewConfig = &oldEnv.Config
	}

	display.Step("Updating environment: %s", opts.OldEnvName)

	display.Step("Updating stack")

	// if force do uninstall + install again
	if opts.Force {
		if err := Delete(DeleteOpts{
			Name:    []string{opts.OldEnvName},
			Context: opts.Context,
		}); err != nil {
			return nil, fmt.Errorf("failed to delete environment: %w", err)
		}
		env, err := Deploy(DeployOpts{
			Context: opts.Context,
			Config:  opts.NewConfig,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to deploy new environment: %w", err)
		}

		return env, nil
	}

	chart, err := config.GetChart()
	if err != nil {
		return nil, fmt.Errorf("TODO: %w", err)
	}

	values, err := opts.NewConfig.AsValues()
	if err != nil {
		return nil, fmt.Errorf("TODO: %w", err)
	}

	settings := cli.New()
	settings.KubeContext = opts.Context
	settings.SetNamespace(opts.NewConfig.Name)

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		opts.NewConfig.Name,
		"secret",
		func(format string, v ...any) { display.Info(format, v...) },
	); err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %w", err)
	}

	client := action.NewUpgrade(actionConfig)
	client.Namespace = opts.NewConfig.Name
	client.Wait = true
	client.WaitForJobs = true
	client.Atomic = true
	client.Timeout = 300 * time.Second

	rel, err := client.Run(opts.NewConfig.Name, chart, values.AsMap())
	if err != nil {
		return nil, fmt.Errorf("failed to install helm chart: %w", err)
	}

	// TODO: populate ontologies

	newEnv, err := ReleaseToEnv(rel, opts.Context)
	if err != nil {
		return nil, fmt.Errorf("TODO: %w", err)
	}

	return newEnv, nil
}

func (u *UpdateOpts) Validate() error {
	if u.OldEnvName == "" {
		return fmt.Errorf("old environment name is required")
	}

	if err := validate.Name(u.OldEnvName); err != nil {
		return fmt.Errorf("'%s' is an invalid name for an environment: %w", u.OldEnvName, err)
	}

	if err := validate.EnvironmentExistsK8s(u.OldEnvName); err != nil {
		return fmt.Errorf("an environment with the name '%s' already exists: %w", u.OldEnvName, err)
	}

	if u.NewConfig != nil {
		if u.Reset {
			return fmt.Errorf("reset and new config are mutually exclusive")
		}

		if err := u.NewConfig.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}

		if !u.Force && u.NewConfig.Name != "" && u.NewConfig.Name != u.OldEnvName {
			// TODO: better error
			return fmt.Errorf("the name in the config is different from the name of the environment to update, this is not supported")
		}

	}

	// TODO: do we really need this?
	contexts, err := common.GetKubeContexts()
	if err != nil {
		return fmt.Errorf("failed to list kubectl contexts: %w", err)
	}
	contextFound := slices.Contains(contexts, u.Context)
	if !contextFound {
		return fmt.Errorf("K8s context %q is not an available context", u.Context)
	}

	return nil
}
