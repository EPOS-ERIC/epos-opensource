package k8score

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/epos-eu/epos-opensource/command"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/db"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/epos-eu/epos-opensource/display"
)

type DeployOpts struct {
	// Optional. path to an env file to use. If not set the default embedded one will be used
	EnvFile string
	// Optional. path to a directory that will contain the custom manifest to use for the deploy
	ManifestDir string
	// Optional. path to a custom directory to store the .env and manifests in
	Path string
	// Required. name of the environment
	Name string
	// Optional. the kubernetes context to be used. uses current kubeclt default if not set
	Context string
	// Optional. the protocol to use for the deployment. currently supports http and https
	Protocol string
	// Optional. custom ip (or hostname) to use instead of using the generated one
	CustomHost string
}

func Deploy(opts DeployOpts) (*sqlc.Kubernetes, error) {
	if opts.Context == "" {
		cmd := exec.Command("kubectl", "config", "current-context")
		out, err := command.RunCommand(cmd, true)
		if err != nil {
			return nil, fmt.Errorf("failed to get current kubectl context: %w", err)
		}
		opts.Context = string(out)
		opts.Context = strings.TrimSpace(opts.Context)
	}

	display.Info("Using kubectl context: %s", opts.Context)
	display.Step("Creating environment: %s", opts.Name)

	host := opts.CustomHost
	if host == "" {
		generatedHost, err := getAPIHost(opts.Context)
		if err != nil {
			return nil, fmt.Errorf("error getting api host: %w", err)
		}
		host = generatedHost
	}

	dir, err := NewEnvDir(opts.EnvFile, opts.ManifestDir, opts.Path, opts.Name, opts.Context, opts.Protocol, host)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare environment directory: %w", err)
	}

	display.Done("Environment created in directory: %s", dir)

	handleFailure := func(msg string, mainErr error) (*sqlc.Kubernetes, error) {
		if err := deleteNamespace(opts.Name, opts.Context); err != nil {
			display.Warn("error deleting namespace %s, %v", opts.Name, err)
		}
		if err := common.RemoveEnvDir(dir); err != nil {
			return nil, fmt.Errorf("error deleting environment %s: %w", dir, err)
		}
		display.Error("stack deployment failed")
		return nil, fmt.Errorf(msg, mainErr)
	}

	if err := deployManifests(dir, opts.Name, true, opts.Context, opts.Protocol); err != nil {
		display.Error("Deploy failed: %v", err)
		return handleFailure("deploy failed: %w", err)
	}

	portalURL, gatewayURL, backofficeURL, err := buildEnvURLs(dir, opts.Context, opts.Protocol, host)
	if err != nil {
		display.Error("error building env urls for the environment: %v", err)
		return handleFailure("error building env urls for environment '%s': %w", fmt.Errorf("%s: %w", dir, err))
	}

	display.Info("Generated URL for data portal: %s", portalURL)
	display.Info("Generated URL for gateway: %s", gatewayURL)
	display.Info("Generated URL for backoffice: %s", backofficeURL)

	display.Step("Starting port-forward to gateway pod for ontologies")

	localPort, err := common.FindFreePort()
	if err != nil {
		return handleFailure("could not find free port: %w", err)
	}

	err = ForwardAndRun(opts.Name, "ingestor-service", localPort, 8080, opts.Context, func(host string, port int) error {
		display.Done("Port forward started successfully")

		url := fmt.Sprintf("http://%s:%d/api/ingestor-service/v1/", host, port)
		if err := common.PopulateOntologies(url); err != nil {
			return fmt.Errorf("error populating ontologies through port-forward: %w", err)
		}
		return nil
	})
	if err != nil {
		display.Error("error initializing the ontologies in the environment: %v", err)
		return handleFailure("error initializing the ontologies: %w", err)
	}

	kube, err := db.InsertKubernetes(opts.Name, dir, opts.Context, gatewayURL, portalURL, backofficeURL, opts.Protocol)
	if err != nil {
		display.Error("failed to insert kubernetes in db: %v", err)
		return handleFailure("failed to insert kubernetes %s (dir: %s) in db: %w", fmt.Errorf("%s, %s, %w", opts.Name, dir, err))
	}

	return kube, nil
}
