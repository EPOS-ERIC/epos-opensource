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

// DeployOpts defines inputs for Deploy.
type DeployOpts struct {
	// Optional. Kubernetes context to use; defaults to the current kubectl context when unset.
	Context string
	// Required. Environment configuration used to build deployment values.
	Config *config.Config
}

// Deploy installs the EPOS Helm chart and returns the deployed environment metadata.
func Deploy(opts DeployOpts) (*Env, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deploy parameters: %w", err)
	}

	display.Step("Deploying environment: %s", opts.Config.Name)

	display.Debug("building urls")

	urls, err := opts.Config.BuildEnvURLs()
	if err != nil {
		display.Error("error building urls: %v", err)
		return nil, err
	}

	display.Debug("built urls: %+v", urls)
	display.Debug("building chart")

	chart, err := config.GetChart()
	if err != nil {
		return nil, fmt.Errorf("failed to load helm chart: %w", err)
	}

	display.Debug("built chart: %+v", chart)
	display.Debug("building values")

	values, err := opts.Config.AsValues()
	if err != nil {
		return nil, fmt.Errorf("failed to build helm values from config: %w", err)
	}

	display.Debug("built values: %+v", values)

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

	display.Debug("running helm install")

	rel, err := client.Run(chart, values.AsMap())
	if err != nil {
		return nil, fmt.Errorf("failed to install helm chart: %w", err)
	}

	display.Debug("installed release: %s (status: %s)", rel.Name, rel.Info.Status)

	env, err := ReleaseToEnv(rel, opts.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to convert release to environment: %w", err)
	}

	display.Done("Deployed environment: %s", opts.Config.Name)

	return env, nil
}

// Validate checks DeployOpts and resolves required deployment preconditions.
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
