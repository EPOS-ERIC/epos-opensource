package dockercore

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/epos-eu/epos-opensource/display"
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
}

func (d *DeployOpts) Deploy() (*sqlc.Docker, error) {
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	display.Step("Creating environment: %s", d.Name)

	dir, err := NewEnvDir(d.EnvFile, d.ComposeFile, d.Path, d.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	display.Done("Environment created in directory: %s", dir)

	var stackDeployed bool
	var cleanupErr error
	cleanup := func() {
		if stackDeployed {
			if derr := downStack(dir, false); derr != nil {
				display.Warn("docker compose down failed, there may be dangling resources: %v", derr)
			}
		}
		if rerr := common.RemoveEnvDir(dir); rerr != nil {
			cleanupErr = fmt.Errorf("error deleting environment %s: %w", dir, rerr)
		}
	}

	if d.PullImages {
		if err := pullEnvImages(dir, d.Name); err != nil {
			display.Error("Pulling images failed: %v", err)
			cleanup()
			if cleanupErr != nil {
				return nil, cleanupErr
			}
			return nil, err
		}
	}

	ports := &DeploymentPorts{
		GUI:        32000,
		API:        33000,
		Backoffice: 34000,
	}
	if d.EnvFile != "" {
		ports, err = loadPortsFromEnvFile(d.EnvFile)
		if err != nil {
			return nil, fmt.Errorf("error loading ports from custom .env file at '%s': %w", d.EnvFile, err)
		}
	} else {
		err := ports.ensureFree()
		if err != nil {
			return nil, fmt.Errorf("failed to find free ports: %w", err)
		}
	}

	urls, err := deployStack(dir, d.Name, ports)
	if err != nil {
		display.Error("Deploy failed: %v", err)
		stackDeployed = true
		cleanup()
		if cleanupErr != nil {
			return nil, cleanupErr
		}
		display.Error("stack deployment failed")
		return nil, err
	}
	stackDeployed = true

	if err := common.PopulateOntologies(urls.apiURL); err != nil {
		display.Error("error initializing the ontologies in the environment: %v", err)
		cleanup()
		if cleanupErr != nil {
			return nil, cleanupErr
		}
		display.Error("stack deployment failed")
		return nil, err
	}

	docker, err := db.InsertDocker(sqlc.Docker{
		Name:           d.Name,
		Directory:      dir,
		ApiUrl:         urls.apiURL,
		GuiUrl:         urls.guiURL,
		BackofficeUrl:  urls.backofficeURL,
		ApiPort:        int64(ports.API),
		GuiPort:        int64(ports.GUI),
		BackofficePort: int64(ports.Backoffice),
	})
	if err != nil {
		cleanup()
		if cleanupErr != nil {
			return nil, cleanupErr
		}
		return nil, fmt.Errorf("failed to insert docker %s (dir: %s) in db: %w", d.Name, dir, err)
	}
	return docker, err
}

// Validate checks if the DeployOpts are valid for a deployment.
// It checks:
// - the name is set
// - the name is not already used by some other environment
// - the path directory (if it exists) does not contain a '.env' or 'docker-compose.yaml' file
func (d *DeployOpts) Validate() error {
	if err := common.ValidateEnvName(d.Name); err != nil {
		return fmt.Errorf("invalid name for environment: %w", err)
	}

	envs, err := db.GetAllDocker()
	if err != nil {
		return fmt.Errorf("error getting installed docker environment from db: %w", err)
	}
	for _, env := range envs {
		if env.Name == d.Name {
			return fmt.Errorf("an environment with the name '%s' already exists", d.Name)
		}
	}

	// path validation
	path, err := filepath.Abs(d.Path)
	if err != nil {
		return fmt.Errorf("error getting absolute path for path: %s", d.Path)
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("cannot stat %q: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%q exists but is not a directory", path)
	}

	forbidden := []string{".env", "docker-compose.yaml"}
	for _, name := range forbidden {
		p := filepath.Join(path, name)
		if _, err := os.Stat(p); err == nil {
			return fmt.Errorf("directory %q already contains %s", path, name)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("error checking for %s: %w", p, err)
		}
	}

	return nil
}
