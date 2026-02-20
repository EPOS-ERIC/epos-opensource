package k8s

import (
	"fmt"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
	"github.com/EPOS-ERIC/epos-opensource/validate"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type DeployOpts struct {
	// Optional. Kubernetes context to use; defaults to the current kubectl context when unset.
	Context string
	// Required. Environment configuration used to build deployment values.
	Config *config.Config
}

func Deploy(opts DeployOpts) (*Env, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deploy parameters: %w", err)
	}

	// TODO: add info/logs

	urls, err := opts.Config.BuildEnvURLs()
	if err != nil {
		display.Error("error building urls: %v", err)
		return nil, err
	}

	display.Debug("urls: %+v", urls)

	chart, err := config.GetChart()
	if err != nil {
		return nil, fmt.Errorf("TODO: %w", err)
	}

	values, err := opts.Config.AsValues()
	if err != nil {
		return nil, fmt.Errorf("TODO: %w", err)
	}

	settings := cli.New()
	settings.KubeContext = opts.Context
	settings.SetNamespace(opts.Config.Name)

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		opts.Config.Name,
		"secret",
		func(format string, v ...any) { display.Info(format, v...) },
	); err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %w", err)
	}

	client := action.NewInstall(actionConfig)
	client.ReleaseName = opts.Config.Name
	client.Namespace = opts.Config.Name
	client.CreateNamespace = true
	client.Wait = true
	client.WaitForJobs = true
	client.Atomic = true
	client.Timeout = 300 * time.Second

	rel, err := client.Run(chart, values.AsMap())
	if err != nil {
		return nil, fmt.Errorf("failed to install helm chart: %w", err)
	}

	display.Debug("Installed release: %s (status: %s)", rel.Name, rel.Info.Status)

	// TODO: populate the ontologies

	env, err := ReleaseToEnv(rel, opts.Context)
	if err != nil {
		return nil, fmt.Errorf("TODO: %w", err)
	}

	return env, nil
}

func (d *DeployOpts) Validate() error {
	display.Debug("context: %s", d.Context)
	display.Debug("config: %+v", d.Config)

	if d.Config == nil {
		return fmt.Errorf("config is required")
	}

	if err := d.Config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if err := validate.Name(d.Config.Name); err != nil {
		return fmt.Errorf("'%s' is an invalid name for an environment: %w", d.Config.Name, err)
	}

	if d.Context == "" {
		context, err := common.GetCurrentKubeContext()
		if err != nil {
			return fmt.Errorf("failed to get current kubectl context: %w", err)
		}

		d.Context = context
	} else if err := EnsureContextExists(d.Context); err != nil {
		return fmt.Errorf("K8s context %q is not an available context: %w", d.Context, err)
	}

	if err := EnsureEnvironmentDoesNotExist(d.Config.Name, d.Context); err != nil {
		return fmt.Errorf("an environment with the name '%s' already exists: %w", d.Config.Name, err)
	}

	return nil
}
