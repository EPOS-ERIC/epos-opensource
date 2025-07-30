package k8score

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/epos-eu/epos-opensource/display"
)

func Deploy(envFile, composeFile, path, name, context, protocol string) (*sqlc.Kubernetes, error) {
	if context == "" {
		cmd := exec.Command("kubectl", "config", "current-context")
		out, err := common.RunCommand(cmd, true)
		if err != nil {
			return nil, fmt.Errorf("failed to get current kubectl context: %w", err)
		}
		context = string(out)
		context = strings.TrimSpace(context)
	}

	display.Info("Using kubectl context: %s", context)
	display.Step("Creating environment: %s", name)

	dir, err := NewEnvDir(envFile, composeFile, path, name, context, protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	display.Done("Environment created in directory: %s", dir)

	handleFailure := func(msg string, mainErr error) (*sqlc.Kubernetes, error) {
		if err := deleteNamespace(name, context); err != nil {
			display.Warn("error deleting namespace %s, %v", name, err)
		}
		if err := common.RemoveEnvDir(dir); err != nil {
			return nil, fmt.Errorf("error deleting environment %s: %w", dir, err)
		}
		display.Error("stack deployment failed")
		return nil, fmt.Errorf(msg, mainErr)
	}

	if err := deployManifests(dir, name, true, context, protocol); err != nil {
		display.Error("Deploy failed: %v", err)
		return handleFailure("deploy failed: %w", err)
	}

	portalURL, gatewayURL, backofficeURL, err := buildEnvURLs(dir, context, protocol)
	if err != nil {
		display.Error("error building env urls for the environment: %v", err)
		return handleFailure("error building env urls for environment '%s': %w", fmt.Errorf("%s: %w", dir, err))
	}

	display.Info("Generated URL for data portal: %s", portalURL)
	display.Info("Generated URL for gateway: %s", gatewayURL)
	display.Info("Generated URL for backoffice: %s", backofficeURL)

	if err := common.PopulateOntologies(gatewayURL); err != nil {
		display.Error("error initializing the ontologies in the environment: %v", err)
		return handleFailure("error initializing the ontologies: %w", err)
	}

	kube, err := db.InsertKubernetes(name, dir, context, gatewayURL, portalURL, backofficeURL, protocol)
	if err != nil {
		display.Error("failed to insert kubernetes in db: %v", err)
		return handleFailure("failed to insert kubernetes %s (dir: %s) in db: %w", fmt.Errorf("%s, %s, %w", name, dir, err))
	}

	return kube, nil
}
