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

	if pullImages {
		if err := pullEnvImages(dir, name); err != nil {
			common.PrintError("Pulling images failed: %v", err)
			if err := common.RemoveEnvDir(dir); err != nil {
				return "", "", fmt.Errorf("error deleting environment %s: %w", dir, err)
			}
			return "", "", err
		}
	}

	if err := deployStack(dir, name); err != nil {
		common.PrintError("Deploy failed: %v", err)

		if err := downStack(dir, false); err != nil {
			common.PrintWarn("docker compose down failed, there may be dangling resources: %v", err)
		}

		if err := common.RemoveEnvDir(dir); err != nil {
			return "", "", fmt.Errorf("error deleting environment %s: %w", dir, err)
		}
		common.PrintError("stack deployment failed")
		return "", "", err
	}

	portalURL, gatewayURL, err = buildEnvURLs(dir)
	if err != nil {
		return "", "", fmt.Errorf("error building env urls for environment '%s': %w", dir, err)
	}

	if err := common.PopulateOntologies(gatewayURL); err != nil {
		common.PrintError("error initializing the ontologies in the environment: %v", err)

		if err := downStack(dir, false); err != nil {
			common.PrintWarn("docker compose down failed, there may be dangling resources: %v", err)
		}

		if err := common.RemoveEnvDir(dir); err != nil {
			return "", "", fmt.Errorf("error deleting environment %s: %w", dir, err)
		}
		common.PrintError("stack deployment failed")
		return "", "", err
	}

	err = db.InsertEnv(name, dir, "docker")
	if err != nil {
		return "", "", fmt.Errorf("failed to insert env %s (dir: %s, platform: %s) in db: %w", name, dir, "docker", err)
	}

	gatewayURL, err = url.JoinPath(gatewayURL, "ui/")
	if err != nil {
		return portalURL, "", fmt.Errorf("failed to build gateway URL: %w", err)
	}
	return portalURL, gatewayURL, err
}
