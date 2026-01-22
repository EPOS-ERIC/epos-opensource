package dockercore

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/EPOS-ERIC/epos-opensource/cmd/docker/dockercore/config"
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/db"
	"github.com/EPOS-ERIC/epos-opensource/db/sqlc"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/validate"
)

type DeployOpts struct {
	// Custom directory path for .env and docker-compose.yaml files
	Path string
	// Pull images before deploying
	PullImages bool
	// Environment configuration (required)
	Config *config.EnvConfig
}

func Deploy(opts DeployOpts) (*sqlc.Docker, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deploy parameters: %w", err)
	}

	if err := opts.Config.EnsurePortsFree(); err != nil {
		return nil, fmt.Errorf("failed to ensure ports are free: %w", err)
	}

	display.Step("Creating environment: %s", opts.Config.Name)

	dir, err := NewEnvDir(opts.Path, opts.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	display.Done("Environment created in directory: %s", dir)

	if !opts.PullImages {
		updates, err := opts.Config.CheckForUpdates()
		if err != nil {
			log.Printf("error checking for updates: %v", err)
		}
		display.ImageUpdatesAvailable(updates, opts.Config.Name)
	}

	var stackDeployed bool
	handleFailure := func(msg string, mainErr error) (*sqlc.Docker, error) {
		if stackDeployed {
			if derr := downStack(dir, false); derr != nil {
				display.Warn("docker compose down failed, there may be dangling resources: %v", derr)
			}
		}
		if err := common.RemoveEnvDir(dir); err != nil {
			return nil, fmt.Errorf("error deleting environment %s: %w", dir, err)
		}
		display.Error("stack deployment failed")
		return nil, fmt.Errorf(msg, mainErr)
	}

	if opts.PullImages {
		if err := pullEnvImages(dir, opts.Config.Name); err != nil {
			display.Error("Pulling images failed: %v", err)
			return handleFailure("pulling images failed: %w", err)
		}
	}

	stackDeployed = true

	err = deployStack(true, dir)
	if err != nil {
		display.Error("Deploy failed: %v", err)
		return handleFailure("deploy failed: %w", err)
	}

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

	var backofficePort int64
	if opts.Config.Components.Backoffice.Enabled {
		backofficePort = int64(opts.Config.Components.Backoffice.GUI.Port)
	}
	docker, err := db.UpsertDocker(sqlc.Docker{
		Name:           opts.Config.Name,
		Directory:      dir,
		ApiUrl:         urls.APIURL,
		GuiUrl:         urls.GUIURL,
		BackofficeUrl:  urls.BackofficeURL,
		ApiPort:        int64(opts.Config.Components.Gateway.Port),
		GuiPort:        int64(opts.Config.Components.PlatformGUI.Port),
		BackofficePort: &backofficePort,
	})
	if err != nil {
		return handleFailure("failed to insert docker in db: %w", fmt.Errorf("%s (dir: %s): %w", opts.Config.Name, dir, err))
	}

	return docker, err
}

func (d *DeployOpts) Validate() error {
	if d.Config == nil {
		return fmt.Errorf("config is required")
	}

	if err := d.Config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if err := validate.Name(d.Config.Name); err != nil {
		return fmt.Errorf("invalid name for environment: %w", err)
	}

	if err := validate.EnvironmentNotExistDocker(d.Config.Name); err != nil {
		return fmt.Errorf("an environment with the name '%s' already exists: %w", d.Config.Name, err)
	}

	if err := validate.PathExists(d.Path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	return nil
}
