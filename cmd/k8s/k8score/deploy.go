package k8score

import (
	"fmt"
	"net/url"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
)

func Deploy(envFile, composeFile, path, name string) (portalURL, gatewayURL string, err error) {
	common.PrintStep("Creating environment: %s", name)

	dir, err := NewEnvDir(envFile, composeFile, path, name)
	if err != nil {
		return "", "", fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	common.PrintDone("Environment created in directory: %s", dir)

	handleFailure := func(msg string, mainErr error) (string, string, error) {
		if err := deleteNamespace(name); err != nil {
			common.PrintWarn("error deleting namespace %s, %v", name, err)
		}
		if err := common.RemoveEnvDir(dir); err != nil {
			return "", "", fmt.Errorf("error deleting environment %s: %w", dir, err)
		}
		common.PrintError("stack deployment failed")
		return "", "", fmt.Errorf(msg, mainErr)
	}

	if err := deployManifests(dir, name, true); err != nil {
		common.PrintError("Deploy failed: %v", err)
		return handleFailure("deploy failed: %w", err)
	}

	portalURL, gatewayURL, err = buildEnvURLs(dir)
	if err != nil {
		common.PrintError("error building env urls for the environment: %v", err)
		return handleFailure("error building env urls for environment '%s': %w", fmt.Errorf("%s: %w", dir, err))
	}

	if err := common.PopulateOntologies(gatewayURL); err != nil {
		common.PrintError("error initializing the ontologies in the environment: %v", err)
		return handleFailure("error initializing the ontologies: %w", err)
	}

	err = db.InsertEnv(name, dir, "kubernetes")
	if err != nil {
		common.PrintError("failed to insert env in db: %v", err)
		return handleFailure("failed to insert env %s (dir: %s, platform: %s) in db: %w", fmt.Errorf("%s, %s, %w", name, dir, err))
	}

	gatewayURL, err = url.JoinPath(gatewayURL, "ui/")
	if err != nil {
		return handleFailure("failed to build gateway URL: %w", err)
	}
	return portalURL, gatewayURL, err
}
