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

	if err := deployStack(dir, name); err != nil {
		common.PrintError("Deploy failed: %v", err)

		// TODO
		// if err := downStack(dir, false); err != nil {
		// 	common.PrintWarn("docker compose down failed, there may be dangling resources: %v", err)
		// }

		if err := common.RemoveEnvDir(dir); err != nil {
			return "", "", fmt.Errorf("error deleting environment %s: %w", dir, err)
		}
		common.PrintError("stack deployment failed")
		return "", "", err
	}
	//
	// if err := populateOntologies(dir); err != nil {
	// 	common.PrintError("error initializing the ontologies in the environment: %v", err)
	//
	// 	if err := downStack(dir, false); err != nil {
	// 		common.PrintWarn("docker compose down failed, there may be dangling resources: %v", err)
	// 	}
	//
	// 	if err := common.RemoveEnvDir(dir); err != nil {
	// 		return "", "", fmt.Errorf("error deleting environment %s: %w", dir, err)
	// 	}
	// 	common.PrintError("stack deployment failed")
	// }
	//
	// portalURL, gatewayURL, err = common.BuildEnvURLs(dir)
	// if err != nil {
	// 	return "", "", fmt.Errorf("error building env urls for environment '%s': %w", dir, err)
	// }

	return common.BuildEnvURLs(dir)
}
