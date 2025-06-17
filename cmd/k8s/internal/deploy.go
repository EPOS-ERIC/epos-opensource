package internal

import (
	"epos-cli/common"
	"fmt"
)

func Deploy(envFile, composeFile, path, name string) (portalURL, gatewayURL string, err error) {
	common.PrintStep("Creating environment: %s", name)

	dir, err := NewEnvDir(envFile, composeFile, path, name)
	if err != nil {
		return "", "", fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	common.PrintDone("Environment created in directory: %s", dir)

	if err := deployManifests(dir, name); err != nil {
		common.PrintError("Deploy failed: %v", err)

		if err := deleteNamespace(name); err != nil {
			common.PrintWarn("error deleting namespace %s, %v", name, err)
		}

		if err := common.RemoveEnvDir(dir); err != nil {
			return "", "", fmt.Errorf("error deleting environment %s: %w", dir, err)
		}
		common.PrintError("stack deployment failed")
		return "", "", err
	}

	if err := common.PopulateOntologies(dir); err != nil {
		common.PrintError("error initializing the ontologies in the environment: %v", err)

		if err := deleteNamespace(name); err != nil {
			common.PrintWarn("error deleting namespace %s, %v", name, err)
		}

		if err := common.RemoveEnvDir(dir); err != nil {
			return "", "", fmt.Errorf("error deleting environment %s: %w", dir, err)
		}
		common.PrintError("stack deployment failed")
	}

	portalURL, gatewayURL, err = common.BuildEnvURLs(dir)
	if err != nil {
		return "", "", fmt.Errorf("error building env urls for environment '%s': %w", dir, err)
	}

	return portalURL, gatewayURL, nil
}
