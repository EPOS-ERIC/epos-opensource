package docker

import (
	"fmt"
	"log"

	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
	"github.com/EPOS-ERIC/epos-opensource/validate"
)

// DeployOpts defines inputs for Deploy.
type DeployOpts struct {
	// Pull images before deploying
	PullImages bool
	// Environment configuration (required)
	Config *config.EnvConfig
}

// Deploy creates and starts a Docker-based EPOS environment and persists it in the local store.
func Deploy(opts DeployOpts) (*Env, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deploy parameters: %w", err)
	}

	display.Debug("ensuring ports are free")

	if err := opts.Config.EnsurePortsFree(); err != nil {
		return nil, fmt.Errorf("failed to ensure ports are free: %w", err)
	}

	display.Step("Deploying environment: %s", opts.Config.Name)

	if !opts.PullImages {
		updates, err := opts.Config.CheckForUpdates()
		if err != nil {
			log.Printf("error checking for updates: %v", err)
		}
		display.ImageUpdatesAvailable(updates, opts.Config.Name)
	}

	var stackDeployed bool
	handleFailure := func(msg string, mainErr error) (*Env, error) {
		if stackDeployed {
			if derr := downStack(opts.Config, false); derr != nil {
				display.Warn("docker compose down failed, there may be dangling resources: %v", derr)
			}
		}

		display.Error("stack deployment failed")
		return nil, fmt.Errorf(msg, mainErr)
	}

	if opts.PullImages {
		display.Debug("pulling images before stack deployment")

		if err := pullEnvImages(opts.Config); err != nil {
			display.Error("Pulling images failed: %v", err)
			return handleFailure("pulling images failed: %w", err)
		}
	}

	stackDeployed = true

	err := deployStack(true, opts.Config)
	if err != nil {
		display.Error("Deploy failed: %v", err)
		return handleFailure("deploy failed: %w", err)
	}

	display.Debug("stack deployed: %s", opts.Config.Name)

	urls, err := opts.Config.BuildEnvURLs()
	if err != nil {
		display.Error("error building urls: %v", err)
		return handleFailure("error building urls: %w", err)
	}

	display.Debug("urls: %+v", urls)

	if err := common.PopulateOntologies(urls.APIURL); err != nil {
		display.Error("error initializing the ontologies in the environment: %v", err)
		return handleFailure("error initializing the ontologies: %w", err)
	}

	display.Debug("initialized base ontologies using: %s", urls.APIURL)

	env, err := upsertEnvConfig(opts.Config)
	if err != nil {
		return handleFailure("failed to persist environment config: %w", err)
	}

	display.Done("Created environment: %s", opts.Config.Name)

	return env, nil
}

// Validate checks DeployOpts and resolves any required preconditions before deployment.
func (d *DeployOpts) Validate() error {
	display.Debug("pullImages: %v", d.PullImages)
	display.Debug("config: %+v", d.Config)

	if d.Config == nil {
		return fmt.Errorf("config is required")
	}

	if err := d.Config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if err := validate.Name(d.Config.Name); err != nil {
		return fmt.Errorf("invalid name for environment: %w", err)
	}

	if err := EnsureEnvironmentDoesNotExist(d.Config.Name); err != nil {
		return fmt.Errorf("an environment with the name '%s' already exists: %w", d.Config.Name, err)
	}

	return nil
}
