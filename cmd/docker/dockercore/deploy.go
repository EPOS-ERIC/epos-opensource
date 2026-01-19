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
	// Optional. path to an env file to use. If not set the default embedded one will be used
	EnvFile string
	// Optional. path to a docker-compose.yaml file. If not set the default embedded one will be used
	ComposeFile string
	// Optional. path to a custom directory to store the .env and docker-compose.yaml in
	Path string
	// Optional. name to give to the environment. If not set it will use the name set in the passed config. If set and also set in the config, this has precedence
	Name string
	// Optional. whether to pull the images before deploying or not
	PullImages bool
	// Optional. config to use for the deployment
	Config *config.EnvConfig
}

func Deploy(opts DeployOpts) (*sqlc.Docker, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deploy parameters: %w", err)
	}
	display.Step("Creating environment: %s", opts.Name)

	cfg := opts.Config
	if cfg == nil {
		cfg = config.GetDefaultConfig()
	}

	if opts.Name != "" {
		cfg.Name = opts.Name
	} else if cfg.Name == "" {
		return nil, fmt.Errorf("environment name is required. Provide it as an argument or in the config file")
	}

	dir, err := NewEnvDir(opts.EnvFile, opts.ComposeFile, opts.Path, opts.Name, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	display.Done("Environment created in directory: %s", dir)

	if !opts.PullImages {
		updates, err := common.CheckEnvForUpdates(cfg.Images)
		if err != nil {
			log.Printf("error checking for updates: %v", err)
		}
		display.ImageUpdatesAvailable(updates, cfg.Name)
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
		if err := pullEnvImages(dir, opts.Name); err != nil {
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

	urls, err := cfg.BuildEnvURLs()
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
	if cfg.Components.Backoffice.Enabled {
		backofficePort = int64(cfg.Components.Backoffice.GUI.Port)
	}
	docker, err := db.UpsertDocker(sqlc.Docker{
		Name:           opts.Name,
		Directory:      dir,
		ApiUrl:         urls.APIURL,
		GuiUrl:         urls.GUIURL,
		BackofficeUrl:  urls.BackofficeURL,
		ApiPort:        int64(cfg.Components.Gateway.Port),
		GuiPort:        int64(cfg.Components.PlatformGUI.Port),
		BackofficePort: &backofficePort,
	})
	if err != nil {
		return handleFailure("failed to insert docker in db: %w", fmt.Errorf("%s (dir: %s): %w", opts.Name, dir, err))
	}

	return docker, err
}

// Validate TODO
func (d *DeployOpts) Validate() error {
	// TODO: also check the cfg.Name if there is
	if err := validate.Name(d.Name); err != nil {
		return fmt.Errorf("invalid name for environment: %w", err)
	}

	if err := validate.EnvironmentNotExistDocker(d.Name); err != nil {
		return fmt.Errorf("an environment with the name '%s' already exists: %w", d.Name, err)
	}

	if err := validate.PathExists(d.Path); err != nil {
		return fmt.Errorf("the path '%s' is not a valid path: %w", d.Path, err)
	}

	if err := validate.IsFile(d.EnvFile); err != nil {
		return fmt.Errorf("the path to .env '%s' is not a file: %w", d.EnvFile, err)
	}

	if err := validate.IsFile(d.ComposeFile); err != nil {
		return fmt.Errorf("the path to docker-compose '%s' is not a file: %w", d.ComposeFile, err)
	}

	return nil
}
