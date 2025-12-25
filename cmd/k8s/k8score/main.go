package k8score

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/epos-eu/epos-opensource/command"
	"github.com/epos-eu/epos-opensource/common"
	"github.com/epos-eu/epos-opensource/display"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

const (
	platform           = "k8s"
	embedManifestsPath = "static/manifests"
)

//go:embed static/manifests
var manifestsFs embed.FS

// EmbeddedManifestContents holds embedded manifest files indexed by filename.
var EmbeddedManifestContents map[string]string

//go:embed static/.env
var EnvFile string

func init() {
	dirEntries, err := fs.ReadDir(manifestsFs, embedManifestsPath)
	if err != nil {
		log.Fatalf("cannot read embedded manifests directory %q: %v", embedManifestsPath, err)
	}

	EmbeddedManifestContents = make(map[string]string, len(dirEntries))

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}

		filePath := path.Join(embedManifestsPath, entry.Name())
		data, err := manifestsFs.ReadFile(filePath)
		if err != nil {
			log.Fatalf("cannot read embedded manifest %q: %v", filePath, err)
		}

		EmbeddedManifestContents[entry.Name()] = string(data)
	}
}

// NewEnvDir creates an environment directory with .env and manifest files.
// Uses custom files if provided, otherwise uses embedded defaults.
// Expands environment variables in all manifest files.
func NewEnvDir(customEnvFilePath, customManifestsDirPath, customPath, name, context, protocol, host string) (string, error) {
	envPath, err := common.BuildEnvPath(customPath, name, platform)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(envPath); err == nil {
		return "", fmt.Errorf("directory %s already exists", envPath)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check directory %s: %w", envPath, err)
	}

	if err := os.MkdirAll(envPath, 0o750); err != nil {
		return "", fmt.Errorf("failed to create env directory %s: %w", envPath, err)
	}

	var ok bool
	defer func() {
		if !ok {
			if removeErr := os.RemoveAll(envPath); removeErr != nil {
				display.Error("Failed to cleanup directory '%s' after error: %v. You may need to remove it manually.", envPath, removeErr)
			}
		}
	}()

	envContent, err := common.GetContentFromPathOrDefault(customEnvFilePath, EnvFile)
	if err != nil {
		return "", fmt.Errorf("failed to get .env file content: %w", err)
	}

	err = os.Setenv("NAMESPACE", name)
	if err != nil {
		return "", fmt.Errorf("failed to set 'NAMESPACE' environment variable: %w", err)
	}
	err = os.Setenv("HOST_NAME", host)
	if err != nil {
		return "", fmt.Errorf("failed to set 'HOST_NAME' environment variable: %w", err)
	}
	expandedEnv := os.ExpandEnv(string(envContent))
	if err := common.CreateFileWithContent(path.Join(envPath, ".env"), expandedEnv); err != nil {
		return "", fmt.Errorf("failed to create .env file: %w", err)
	}

	if customManifestsDirPath != "" {
		if err := copyCustomManifests(customManifestsDirPath, envPath); err != nil {
			return "", fmt.Errorf("failed to copy custom manifests: %w", err)
		}
	} else {
		if err := copyEmbeddedManifests(envPath); err != nil {
			return "", fmt.Errorf("failed to copy embedded manifests: %w", err)
		}
	}

	if err := loadEnvAndExpandManifests(envPath, name, protocol, host); err != nil {
		return "", fmt.Errorf("failed to process environment variables: %w", err)
	}

	ok = true
	return envPath, nil
}

func copyCustomManifests(customManifestsDirPath, envPath string) error {
	dirEntries, err := os.ReadDir(customManifestsDirPath)
	if err != nil {
		return fmt.Errorf("error reading manifests directory %q: %w", customManifestsDirPath, err)
	}

	for _, de := range dirEntries {
		if de.IsDir() || (filepath.Ext(de.Name()) != ".yaml" && filepath.Ext(de.Name()) != ".yml") {
			display.Warn("Found unknown item %s in custom manifest dir, ignoring it", de.Name())
			continue
		}

		srcPath := filepath.Join(customManifestsDirPath, de.Name())
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("error reading file %q: %w", de.Name(), err)
		}

		destPath := filepath.Join(envPath, de.Name())
		if err := common.CreateFileWithContent(destPath, string(data)); err != nil {
			return fmt.Errorf("failed to create manifest file %q: %w", de.Name(), err)
		}
	}

	return nil
}

func copyEmbeddedManifests(envPath string) error {
	for filename, content := range EmbeddedManifestContents {
		destPath := filepath.Join(envPath, filename)
		if err := common.CreateFileWithContent(destPath, content); err != nil {
			return fmt.Errorf("failed to create embedded manifest %q: %w", filename, err)
		}
	}

	return nil
}

// getAPIHost returns the public host for the environment.
// If HOST_NAME is set in the .env file, it is returned.
// Otherwise, it attempts to get the ingress controller IP/hostname, and if that fails, falls back to the local IP.
func getAPIHost(context string) (string, error) {
	ip, err := getIngressControllerIP(context)
	if err != nil {
		display.Warn("error getting ingress IP, falling back to local IP: %v", err)
		ip, err = common.GetLocalIP()
		if err != nil {
			return "", fmt.Errorf("error getting IP address: %w", err)
		}
	}
	return ip, nil
}

func loadEnvAndExpandManifests(envPath, name, protocol, host string) error {
	envFilePath := path.Join(envPath, ".env")
	if err := godotenv.Load(envFilePath); err != nil {
		return fmt.Errorf("failed to load environment file %q: %w", envFilePath, err)
	}

	err := os.Setenv("NAMESPACE", name)
	if err != nil {
		return fmt.Errorf("failed to set 'NAMESPACE' environment variable: %w", err)
	}
	_, apiURL, _, err := buildEnvURLs(envPath, protocol, host)
	if err != nil {
		return fmt.Errorf("error building API URL: %w", err)
	}
	err = os.Setenv("API_HOST", apiURL)
	if err != nil {
		return fmt.Errorf("failed to set 'API_HOST' environment variable: %w", err)
	}

	files, err := os.ReadDir(envPath)
	if err != nil {
		return fmt.Errorf("failed to read environment directory %q: %w", envPath, err)
	}

	for _, de := range files {
		if de.IsDir() || de.Name() == ".env" {
			continue
		}

		filePath := filepath.Join(envPath, de.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", de.Name(), err)
		}

		expanded := os.ExpandEnv(string(data))
		if err := common.CreateFileWithContent(filePath, expanded); err != nil {
			return fmt.Errorf("failed to write expanded manifest file %q: %w", de.Name(), err)
		}
	}

	return nil
}

// runKubectl executes kubectl with given args in the specified directory.
func runKubectl(dir string, suppressOut bool, context string, args ...string) error {
	if context != "" {
		args = append(args, "--context", context)
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Dir = dir
	_, err := command.RunCommand(cmd, suppressOut)
	return err
}

// applyParallel runs kubectl apply for all targets concurrently.
// If withService is true, applies both deployment-X.yaml and service-X.yaml files.
func applyParallel(dir string, targets []string, withService bool, context string) error {
	var g errgroup.Group

	for _, t := range targets {
		g.Go(func() error {
			if withService {
				return runKubectl(dir, false, context,
					"apply", "-f", fmt.Sprintf("deployment-%s.yaml", t), "-f", fmt.Sprintf("service-%s.yaml", t))
			}
			return runKubectl(dir, false, context, "apply", "-f", t)
		})
	}
	return g.Wait()
}

// waitDeployments waits for all deployments to be ready with 1 minute timeout each.
func waitDeployments(dir, namespace string, names []string, context string) error {
	var g errgroup.Group

	for _, n := range names {
		g.Go(func() error {
			return runKubectl(dir, false, context,
				"rollout", "status", fmt.Sprintf("deployment/%s", n),
				"--timeout", (10 * time.Minute).String(),
				"-n", namespace)
		})
	}
	return g.Wait()
}

// deployManifests deploys all resources to the namespace in stages:
// 1. Create namespace, 2. Setup environment, 3. Deploy infra, 4. Deploy services, 5. Deploy gateway/portal.
func deployManifests(dir, namespace string, createNamespace bool, context, protocol string) error {
	if createNamespace {
		display.Step("Creating namespace %s", namespace)
		if err := runKubectl(dir, true, context, "get", "namespace", namespace); err != nil {
			if err := runKubectl(dir, false, context, "create", "namespace", namespace); err != nil {
				return fmt.Errorf("failed to create namespace %s: %w", namespace, err)
			}
		} else {
			return fmt.Errorf("namespace %s already exists", namespace)
		}
	}

	display.Step("Setting up the environment")
	setup := []string{
		"configmap-epos-env.yaml",
		"pvc-psqldata.yaml",
		"pvc-converter-plugins.yaml",
	}
	for _, f := range setup {
		if err := runKubectl(dir, false, context, "apply", "-f", f); err != nil {
			return fmt.Errorf("failed to apply %s: %w", f, err)
		}
	}

	infra := []string{
		"rabbitmq",
		"metadata-database",
	}
	display.Step("Deploying infrastructure components")
	if err := applyParallel(dir, infra, true, context); err != nil {
		return fmt.Errorf("failed to deploy infrastructure components: %w", err)
	}
	if err := runKubectl(dir, false, context, "apply", "-f", "service-rabbitmq-management.yaml"); err != nil {
		return fmt.Errorf("failed to apply RabbitMQ management service: %w", err)
	}

	display.Step("Waiting for infrastructure to be ready...")
	if err := waitDeployments(dir, namespace, infra, context); err != nil {
		return fmt.Errorf("infrastructure deployment failed: %w", err)
	}

	services := []string{
		"resources-service",
		"ingestor-service",
		"external-access-service",
		"converter-service",
		"converter-routine",
		"backoffice-service",
		"email-sender-service",
		"sharing-service",
	}
	display.Step("Deploying services")
	if err := applyParallel(dir, services, true, context); err != nil {
		return fmt.Errorf("failed to deploy services: %w", err)
	}

	display.Step("Waiting for services to be ready...")
	if err := waitDeployments(dir, namespace, services, context); err != nil {
		return fmt.Errorf("services deployment failed: %w", err)
	}

	topLevel := []string{
		"gateway",
		"dataportal",
		"backoffice-ui",
	}
	display.Step("Deploying top-level services")
	if err := applyParallel(dir, topLevel, true, context); err != nil {
		return fmt.Errorf("failed to deploy gateway and dataportal: %w", err)
	}

	display.Step("Waiting for top-level services to be ready...")
	if err := waitDeployments(dir, namespace, topLevel, context); err != nil {
		return fmt.Errorf("gateway and dataportal deployment failed: %w", err)
	}

	// deploy the ingresses *after* the top-level services
	var ingresses string
	switch protocol {
	case "https":
		ingresses = "ingresses-secure.yaml"
	case "http":
		ingresses = "ingresses-insecure.yaml"
	default:
		return fmt.Errorf("unknown protocol %s", protocol)
	}

	display.Step("Deploying ingresses")
	if err := runKubectl(dir, false, context, "apply", "-f", ingresses); err != nil {
		return fmt.Errorf("failed to apply %s: %w", ingresses, err)
	}
	if err := waitIngresses(context, []string{"gateway", "dataportal", "backoffice-ui"}, namespace); err != nil {
		return fmt.Errorf("failed to wait for ingresses to be ready: %w", err)
	}

	return nil
}

// deleteNamespace removes the specified namespace and all its resources.
func deleteNamespace(name, context string) error {
	if err := runKubectl(".", false, context, "delete", "namespace", name); err != nil {
		return fmt.Errorf("failed to delete namespace %s: %w", name, err)
	}
	return nil
}

// buildEnvURLs returns the base urls for the dataportal and the gateway.
// It tries to get the ip of the ingress of the cluster, and if there is an error it returns the localIP
func buildEnvURLs(dir, protocol, host string) (portalURL, gatewayURL, backofficeURL string, err error) {
	env, err := godotenv.Read(filepath.Join(dir, ".env"))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read .env file at %s: %w", filepath.Join(dir, ".env"), err)
	}

	apiDeployPath, ok := env["API_DEPLOY_PATH"]
	if !ok {
		return "", "", "", fmt.Errorf("environment variable API_DEPLOY_PATH is not set")
	}
	dataportalDeployPath, ok := env["DATAPORTAL_DEPLOY_PATH"]
	if !ok {
		return "", "", "", fmt.Errorf("environment variable DATAPORTAL_DEPLOY_PATH is not set")
	}
	backofficeDeployPath, ok := env["BACKOFFICE_DEPLOY_PATH"]
	if !ok {
		return "", "", "", fmt.Errorf("environment variable BACKOFFICE_DEPLOY_PATH is not set")
	}

	gatewayURL, err = url.JoinPath(fmt.Sprintf("%s://%s", protocol, host), apiDeployPath)
	if err != nil {
		return "", "", "", fmt.Errorf("error building gateway url: %w", err)
	}
	// NOTE: this is needed because the java services expect the api url to not have any trailing slashes. Last check 4.7.25
	gatewayURL = strings.Trim(gatewayURL, "/")

	portalURL, err = url.JoinPath(fmt.Sprintf("%s://%s", protocol, host), dataportalDeployPath)
	if err != nil {
		return "", "", "", fmt.Errorf("error building dataportal url: %w", err)
	}

	backofficeURL, err = url.JoinPath(fmt.Sprintf("%s://%s", protocol, host), backofficeDeployPath)
	if err != nil {
		return "", "", "", fmt.Errorf("error building dataportal url: %w", err)
	}

	return portalURL, gatewayURL, backofficeURL, nil
}

// getIngressControllerIP returns the ingress controller IP or hostname for the context
func getIngressControllerIP(context string) (string, error) {
	if context == "" {
		return "", fmt.Errorf("context must be provided and cannot be empty")
	}
	const (
		maxWait  = 30 * time.Second
		interval = 2 * time.Second
	)
	fields := [][2]string{
		{"hostname", "{.status.loadBalancer.ingress[0].hostname}"},
		{"ip", "{.status.loadBalancer.ingress[0].ip}"},
	}

	display.Step("Waiting for ingress IP/hostname to be assigned...")
	start := time.Now()
	var lastErr error

	for time.Since(start) < maxWait {
		for _, field := range fields {
			desc := field[0]
			jsonpath := field[1]
			args := []string{"get", "service", "ingress-nginx-controller", "-n", "ingress-nginx", "-o", "jsonpath=" + jsonpath, "--context", context}
			cmd := exec.Command("kubectl", args...)
			out, err := command.RunCommand(cmd, true)
			if err != nil {
				lastErr = fmt.Errorf("error getting ingress %s: %w", desc, err)
				continue
			}
			value := strings.TrimSpace(out)
			if value != "" {
				if desc == "hostname" && value == "localhost" {
					localIP, err := common.GetLocalIP()
					if err != nil {
						return "", fmt.Errorf("error getting local IP: %w", err)
					}
					value = localIP
				}
				display.Done("Ingress assigned %s: %s", desc, value)
				return value, nil
			}
		}
		time.Sleep(interval)
	}
	if lastErr != nil {
		return "", fmt.Errorf("error waiting for ingress to be ready, last error: %w", lastErr)
	}
	return "", fmt.Errorf("error getting ingress IP: both ip and hostname are empty after waiting %s", maxWait)
}

func waitIngresses(kubeContext string, ingressNames []string, namespace string) error {
	const (
		maxWait  = 2 * time.Minute
		interval = 5 * time.Second
	)

	display.Step("Waiting for ingresses to be ready...")

	var g errgroup.Group

	for _, ingressName := range ingressNames {
		g.Go(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), maxWait)
			defer cancel()

			for {
				args := []string{
					"get", "ingress", ingressName,
					"-n", namespace,
					"-o", "jsonpath={.status.loadBalancer.ingress[0]}",
					"--context", kubeContext,
				}
				cmd := exec.Command("kubectl", args...)
				out, err := command.RunCommand(cmd, true)
				if err == nil {
					value := strings.TrimSpace(out)
					if value != "" && value != "{}" {
						display.Done("Ingress %s is ready", ingressName)
						return nil
					}
				}

				select {
				case <-ctx.Done():
					return fmt.Errorf("timeout waiting for ingress %s", ingressName)
				case <-time.After(interval):
					continue
				}
			}
		})
	}

	err := g.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait for ingrsesses to be ready: %w", err)
	}
	display.Done("All ingresses are ready")
	return nil
}
