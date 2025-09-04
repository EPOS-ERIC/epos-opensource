package dockercore

import (
	_ "embed"
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/epos-eu/epos-opensource/display"
	"github.com/epos-eu/epos-opensource/validate"
)

type DeployOpts struct {
	// Optional. path to an env file to use. If not set the default embedded one will be used
	EnvFile string
	// Optional. path to a docker-compose.yaml file. If not set the default embedded one will be used
	ComposeFile string
	// Optional. path to a custom directory to store the .env and docker-compose.yaml in
	Path string
	// Required. name of the environment
	Name string
	// Optional. whether to pull the images before deploying or not
	PullImages bool
	// Optional. custom ip to use instead of localhost if set
	CustomHost string
}

func Deploy(opts DeployOpts) (*sqlc.Docker, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deploy parameters: %w", err)
	}
	display.Step("Creating environment: %s", opts.Name)

	dir, err := NewEnvDir(opts.EnvFile, opts.ComposeFile, opts.Path, opts.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	display.Done("Environment created in directory: %s", dir)

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

	ports := &DeploymentPorts{
		GUI:        32000,
		API:        33000,
		Backoffice: 34000,
	}
	if opts.EnvFile != "" {
		ports, err = loadPortsFromEnvFile(dir)
		if err != nil {
			return handleFailure("error loading ports from custom .env file at %w", fmt.Errorf("%s: %w", opts.EnvFile, err))
		}
	} else {
		err := ports.ensureFree()
		if err != nil {
			return handleFailure("failed to find free ports: %w", err)
		}
	}

	urls, err := deployStack(dir, opts.Name, ports, opts.CustomHost)
	if err != nil {
		display.Error("Deploy failed: %v", err)
		stackDeployed = true
		return handleFailure("deploy failed: %w", err)
	}
	stackDeployed = true

	if err := common.PopulateOntologies(urls.apiURL); err != nil {
		display.Error("error initializing the ontologies in the environment: %v", err)
		return handleFailure("error initializing the ontologies: %w", err)
	}

	docker, err := db.InsertDocker(sqlc.Docker{
		Name:           opts.Name,
		Directory:      dir,
		ApiUrl:         urls.apiURL,
		GuiUrl:         urls.guiURL,
		BackofficeUrl:  urls.backofficeURL,
		ApiPort:        int64(ports.API),
		GuiPort:        int64(ports.GUI),
		BackofficePort: int64(ports.Backoffice),
	})
	if err != nil {
		return handleFailure("failed to insert docker in db: %w", fmt.Errorf("%s (dir: %s): %w", opts.Name, dir, err))
	}
	return docker, err
}
func (d *DeployOpts) Validate() error {
	if err := validate.Name(d.Name); err != nil {
		return fmt.Errorf("invalid name for environment: %w", err)
	}
	if err := validate.EnvironmentNotExistDocker(d.Name); err != nil {
		return fmt.Errorf("an environment with the name '%s' already exists: %w", d.Name, err)
	}
	if err := validate.PathExists(d.Path); err != nil {
		return fmt.Errorf("the path '%s' is not a valid path: %w", d.Path, err)
	}
	if err := validate.CustomHost(d.CustomHost); err != nil {
		return fmt.Errorf("the custom host '%s' is not a valid ip or hostname: %w", d.CustomHost, err)
	}
	if err := validate.IsFile(d.EnvFile); err != nil {
		return fmt.Errorf("the path to .env '%s' is not a file: %w", d.EnvFile, err)
	}
	if err := validate.IsFile(d.ComposeFile); err != nil {
		return fmt.Errorf("the path to docker-compose '%s' is not a file: %w", d.ComposeFile, err)
	}
	return nil
}
