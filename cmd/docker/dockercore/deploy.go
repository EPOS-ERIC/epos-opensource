package dockercore

import (
	_ "embed"
	"fmt"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/display"
)

func Deploy(envFile, composeFile, path, name string, pullImages bool) (*db.Docker, error) {
	display.Step("Creating environment: %s", name)

	dir, err := NewEnvDir(envFile, composeFile, path, name)
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

	if pullImages {
		if err := pullEnvImages(dir, name); err != nil {
			display.Error("Pulling images failed: %v", err)
			cleanup()
			if cleanupErr != nil {
				return nil, cleanupErr
			}
			return nil, err
		}
	}

	if err := deployStack(dir, name); err != nil {
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

	portalURL, gatewayURL, backofficeURL, err := buildEnvURLs(dir)
	if err != nil {
		cleanup()
		if cleanupErr != nil {
			return nil, cleanupErr
		}
		return nil, fmt.Errorf("error building env urls for environment '%s': %w", dir, err)
	}

	if err := common.PopulateOntologies(gatewayURL); err != nil {
		display.Error("error initializing the ontologies in the environment: %v", err)
		cleanup()
		if cleanupErr != nil {
			return nil, cleanupErr
		}
		display.Error("stack deployment failed")
		return nil, err
	}

	docker, err := db.InsertDocker(name, dir, gatewayURL, portalURL, backofficeURL)
	if err != nil {
		cleanup()
		if cleanupErr != nil {
			return nil, cleanupErr
		}
		return nil, fmt.Errorf("failed to insert docker %s (dir: %s) in db: %w", name, dir, err)
	}
	return docker, err
}
