package k8score

import (
	"fmt"
	"slices"

	"github.com/EPOS-ERIC/epos-opensource/cmd/k8s/k8score/config"
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/db/sqlc"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/validate"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type DeployOpts struct {
	// Optional. TODO
	Path string
	// Optional. the K8s context to be used. uses current kubeclt default if not set
	Context string
	// Environment configuration (required)
	Config *config.EnvConfig
}

func Deploy(opts DeployOpts) (*sqlc.K8s, error) {
	// TODO: do we really need this context fallback?
	if opts.Context == "" {
		ctx, err := common.GetCurrentKubeContext()
		if err != nil {
			return nil, fmt.Errorf("failed to get current kubectl context: %w", err)
		}
		opts.Context = ctx
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deploy parameters: %w", err)
	}

	// TODO: add info/logs

	// TODO: handle failure?

	urls, err := opts.Config.BuildEnvURLs()
	if err != nil {
		display.Error("error building urls: %v", err)
		return nil, nil
		// return handleFailure("error building urls: %w", err)
	}

	display.Debug("urls: %+v", urls)

	// TODO: deploy using helm sdk

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

	rel, err := client.Run(chart, values.AsMap())
	if err != nil {
		return nil, fmt.Errorf("failed to install helm chart: %w", err)
	}

	display.Debug("Installed release: %s (status: %s)", rel.Name, rel.Info.Status)

	return nil, nil
}

func (d *DeployOpts) Validate() error {
	if d.Config == nil {
		return fmt.Errorf("config is required")
	}

	if err := d.Config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if err := validate.Name(d.Config.Name); err != nil {
		return fmt.Errorf("'%s' is an invalid name for an environment: %w", d.Config.Name, err)
	}

	if err := validate.EnvironmentNotExistK8s(d.Config.Name); err != nil {
		return fmt.Errorf("an environment with the name '%s' already exists: %w", d.Config.Name, err)
	}

	if err := validate.PathExists(d.Path); err != nil {
		return fmt.Errorf("the path '%s' is not a valid path: %w", d.Path, err)
	}

	// TODO: do we really need this?
	contexts, err := common.GetKubeContexts()
	if err != nil {
		return fmt.Errorf("failed to list kubectl contexts: %w", err)
	}
	contextFound := slices.Contains(contexts, d.Context)
	if !contextFound {
		return fmt.Errorf("K8s context %q is not an available context", d.Context)
	}

	return nil
}
