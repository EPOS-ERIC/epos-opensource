package dockercore

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
)

func Deploy(envFile, composeFile, path, name string, pullImages bool) (portalURL, gatewayURL string, err error) {
	common.PrintStep("Creating environment: %s", name)

	dir, err := NewEnvDir(envFile, composeFile, path, name)
	if err != nil {
		return "", "", fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	common.PrintDone("Environment created in directory: %s", dir)

	var stackDeployed bool
	var cleanupErr error
	cleanup := func() {
		if stackDeployed {
			if derr := downStack(dir, false); derr != nil {
				common.PrintWarn("docker compose down failed, there may be dangling resources: %v", derr)
			}
		}
		if rerr := common.RemoveEnvDir(dir); rerr != nil {
			cleanupErr = fmt.Errorf("error deleting environment %s: %w", dir, rerr)
		}
	}

	if pullImages {
		if err := pullEnvImages(dir, name); err != nil {
			common.PrintError("Pulling images failed: %v", err)
			cleanup()
			if cleanupErr != nil {
				return "", "", cleanupErr
			}
			return "", "", err
		}
	}

	if err := deployStack(dir, name); err != nil {
		common.PrintError("Deploy failed: %v", err)
		stackDeployed = true
		cleanup()
		if cleanupErr != nil {
			return "", "", cleanupErr
		}
		common.PrintError("stack deployment failed")
		return "", "", err
	}
	stackDeployed = true

	portalURL, gatewayURL, err = buildEnvURLs(dir)
	if err != nil {
		cleanup()
		if cleanupErr != nil {
			return "", "", cleanupErr
		}
		return "", "", fmt.Errorf("error building env urls for environment '%s': %w", dir, err)
	}

	if err := common.PopulateOntologies(gatewayURL); err != nil {
		common.PrintError("error initializing the ontologies in the environment: %v", err)
		cleanup()
		if cleanupErr != nil {
			return "", "", cleanupErr
		}
		common.PrintError("stack deployment failed")
		return "", "", err
	}

	err = db.InsertEnv(name, dir, "docker")
	if err != nil {
		cleanup()
		if cleanupErr != nil {
			return "", "", cleanupErr
		}
		return "", "", fmt.Errorf("failed to insert env %s (dir: %s, platform: %s) in db: %w", name, dir, "docker", err)
	}

	gatewayURL, err = url.JoinPath(gatewayURL, "ui/")
	if err != nil {
		cleanup()
		if cleanupErr != nil {
			return "", "", cleanupErr
		}
		return portalURL, "", fmt.Errorf("failed to build gateway URL: %w", err)
	}
	return portalURL, gatewayURL, err
}
