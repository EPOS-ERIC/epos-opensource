// Package k8s contains the internal functions used by the k8s cmd to manage the environments
package k8s

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
	"github.com/EPOS-ERIC/epos-opensource/validate"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// UpdateOpts defines inputs for Update.
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

// Update updates an existing K8s environment in place or by full reinstall when Force is enabled.
func Update(opts UpdateOpts) (*Env, error) {
	display.Debug("oldEnvName: %s", opts.OldEnvName)
	display.Debug("context: %s", opts.Context)
	display.Debug("force: %v", opts.Force)
	display.Debug("reset: %v", opts.Reset)
	display.Debug("newConfig: %+v", opts.NewConfig)

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters for update command: %w", err)
	}

	if opts.NewConfig == nil && opts.Reset {
		display.Info("Reset flag enabled: using default config")

		opts.NewConfig = config.GetDefaultConfig()
	}

	if opts.NewConfig != nil && opts.NewConfig.Name == "" {
		opts.NewConfig.Name = opts.OldEnvName

		display.Debug("new config name not set, using old name: %s", opts.NewConfig.Name)
	}

	if opts.NewConfig == nil && !opts.Reset {
		display.Debug("new config not provided, using current environment config")

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
		display.Info("Force flag enabled: reinstalling environment")

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

		display.Done("Updated environment: %s", opts.OldEnvName)

		return env, nil
	}

	display.Debug("building chart")

	chart, err := config.GetChart()
	if err != nil {
		return nil, fmt.Errorf("TODO: %w", err)
	}

	display.Debug("built chart: %+v", chart)
	display.Debug("building values")

	values, err := opts.NewConfig.AsValues()
	if err != nil {
		return nil, fmt.Errorf("TODO: %w", err)
	}

	display.Debug("built values: %+v", values)

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

	display.Debug("running helm upgrade")

	rel, err := client.Run(opts.NewConfig.Name, chart, values.AsMap())
	if err != nil {
		return nil, fmt.Errorf("failed to install helm chart: %w", err)
	}

	display.Debug("upgraded release: %s (status: %s)", rel.Name, rel.Info.Status)

	newEnv, err := ReleaseToEnv(rel, opts.Context)
	if err != nil {
		return nil, fmt.Errorf("TODO: %w", err)
	}

	display.Done("Updated environment: %s", opts.OldEnvName)

	return newEnv, nil
}

// Validate checks UpdateOpts and validates update semantics against existing environment state.
func (u *UpdateOpts) Validate() error {
	display.Debug("oldEnvName: %s", u.OldEnvName)
	display.Debug("context: %s", u.Context)
	display.Debug("force: %v", u.Force)
	display.Debug("reset: %v", u.Reset)
	display.Debug("newConfig: %+v", u.NewConfig)

	if u.OldEnvName == "" {
		return fmt.Errorf("old environment name is required")
	}

	if err := validate.Name(u.OldEnvName); err != nil {
		return fmt.Errorf("'%s' is an invalid name for an environment: %w", u.OldEnvName, err)
	}

	if u.NewConfig != nil {
		if u.Reset {
			return fmt.Errorf("reset and new config are mutually exclusive")
		}

		if u.NewConfig.Name != "" && u.NewConfig.Name != u.OldEnvName {
			return fmt.Errorf("config name %q must match environment name %q", u.NewConfig.Name, u.OldEnvName)
		}

		if err := u.NewConfig.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}

	}

	if u.Context == "" {
		context, err := common.GetCurrentKubeContext()
		if err != nil {
			return fmt.Errorf("failed to get current kubectl context: %w", err)
		}

		u.Context = context

		display.Debug("using current kubectl context: %s", u.Context)
	} else if err := EnsureContextExists(u.Context); err != nil {
		return fmt.Errorf("K8s context %q is not an available context: %w", u.Context, err)
	}

	if err := EnsureEnvironmentExists(u.OldEnvName, u.Context); err != nil {
		return fmt.Errorf("no environment with the name '%s' exists: %w", u.OldEnvName, err)
	}

	return nil
}
