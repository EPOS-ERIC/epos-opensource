package internal

import (
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

	"github.com/epos-eu/epos-opensource/common"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

const (
	pathPrefix         = "k8s"
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
func NewEnvDir(customEnvFilePath, customManifestsDirPath, customPath, name string) (string, error) {
	envPath := common.BuildEnvPath(customPath, name, pathPrefix)

	if _, err := os.Stat(envPath); err == nil {
		return "", fmt.Errorf("directory %s already exists", envPath)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check directory %s: %w", envPath, err)
	}

	if err := os.MkdirAll(envPath, 0777); err != nil {
		return "", fmt.Errorf("failed to create env directory %s: %w", envPath, err)
	}

	var ok bool
	defer func() {
		if !ok {
			if removeErr := os.RemoveAll(envPath); removeErr != nil {
				common.PrintError("Failed to cleanup directory '%s' after error: %v. You may need to remove it manually.", envPath, removeErr)
			}
		}
	}()

	envContent, err := common.GetContentFromPathOrDefault(customEnvFilePath, EnvFile)
	if err != nil {
		return "", fmt.Errorf("failed to get .env file content: %w", err)
	}

	if err := common.CreateFileWithContent(path.Join(envPath, ".env"), envContent); err != nil {
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

	if err := loadEnvAndExpandManifests(envPath, name); err != nil {
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
		if de.IsDir() {
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

func loadEnvAndExpandManifests(envPath, name string) error {
	envFilePath := path.Join(envPath, ".env")
	if err := godotenv.Load(envFilePath); err != nil {
		return fmt.Errorf("failed to load environment file %q: %w", envFilePath, err)
	}

	os.Setenv("NAMESPACE", name)
	_, apiURL, err := buildEnvURLs(envPath)
	if err != nil {
		return fmt.Errorf("error building API URL: %w", err)
	}
	os.Setenv("API_HOST", apiURL)

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
func runKubectl(dir string, suppressOut bool, args ...string) error {
	cmd := exec.Command("kubectl", args...)
	cmd.Dir = dir
	_, err := common.RunCommand(cmd, suppressOut)
	return err
}

// applyParallel runs kubectl apply for all targets concurrently.
// If withService is true, applies both deployment-X.yaml and service-X.yaml files.
func applyParallel(dir string, targets []string, withService bool) error {
	var g errgroup.Group

	for _, t := range targets {
		g.Go(func() error {
			if withService {
				return runKubectl(dir, false,
					"apply", "-f", fmt.Sprintf("deployment-%s.yaml", t), "-f", fmt.Sprintf("service-%s.yaml", t))
			}
			return runKubectl(dir, false, "apply", "-f", t)
		})
	}
	return g.Wait()
}

// waitDeployments waits for all deployments to be ready with 1 minute timeout each.
func waitDeployments(dir, namespace string, names []string) error {
	var g errgroup.Group

	for _, n := range names {
		g.Go(func() error {
			return runKubectl(dir, false,
				"rollout", "status", fmt.Sprintf("deployment/%s", n),
				"--timeout", (2 * time.Minute).String(),
				"-n", namespace)
		})
	}
	return g.Wait()
}

// deployManifests deploys all resources to the namespace in stages:
// 1. Create namespace, 2. Setup environment, 3. Deploy infra, 4. Deploy services, 5. Deploy gateway/portal.
func deployManifests(dir, namespace string, createNamespace bool) error {
	if createNamespace {
		common.PrintStep("Creating namespace %s", namespace)
		if err := runKubectl(dir, true, "get", "namespace", namespace); err != nil {
			if err := runKubectl(dir, false, "create", "namespace", namespace); err != nil {
				return fmt.Errorf("failed to create namespace %s: %w", namespace, err)
			}
		} else {
			return fmt.Errorf("namespace %s already exists", namespace)
		}
	}

	common.PrintStep("Setting up the environment")
	setup := []string{
		"configmap-epos-env.yaml",
		"secret-epos-secret.yaml",
		"pvc-psqldata.yaml",
		"pvc-converter-plugins.yaml",
	}
	for _, f := range setup {
		if err := runKubectl(dir, false, "apply", "-f", f); err != nil {
			return fmt.Errorf("failed to apply %s: %w", f, err)
		}
	}

	infra := []string{"rabbitmq", "metadata-database"}
	common.PrintStep("Deploying infrastructure components")
	if err := applyParallel(dir, infra, true); err != nil {
		return fmt.Errorf("failed to deploy infrastructure components: %w", err)
	}
	if err := runKubectl(dir, false, "apply", "-f", "service-rabbitmq-management.yaml"); err != nil {
		return fmt.Errorf("failed to apply RabbitMQ management service: %w", err)
	}

	common.PrintStep("Waiting for infrastructure to be ready")
	if err := waitDeployments(dir, namespace, infra); err != nil {
		return fmt.Errorf("infrastructure deployment failed: %w", err)
	}

	services := []string{
		"resources-service",
		"ingestor-service",
		"external-access-service",
		"converter-service",
		"converter-routine",
		"backoffice-service",
	}
	common.PrintStep("Deploying services")
	if err := applyParallel(dir, services, true); err != nil {
		return fmt.Errorf("failed to deploy services: %w", err)
	}

	common.PrintStep("Waiting for services to be ready")
	if err := waitDeployments(dir, namespace, services); err != nil {
		return fmt.Errorf("services deployment failed: %w", err)
	}

	finals := []string{"gateway", "dataportal"}
	common.PrintStep("Deploying gateway and dataportal")
	if err := applyParallel(dir, finals, true); err != nil {
		return fmt.Errorf("failed to deploy gateway and dataportal: %w", err)
	}
	if err := waitDeployments(dir, namespace, finals); err != nil {
		return fmt.Errorf("gateway and dataportal deployment failed: %w", err)
	}

	return nil
}

// deleteNamespace removes the specified namespace and all its resources.
func deleteNamespace(name string) error {
	if err := runKubectl(".", false, "delete", "namespace", name); err != nil {
		return fmt.Errorf("failed to delete namespace %s: %w", name, err)
	}
	return nil
}

// buildEnvURLs returns the base urls for the dataportal and the gateway.
// It tries to get the ip of the ingress of the cluster, and if there is an error it returns the localIP
func buildEnvURLs(dir string) (portalURL, gatewayURL string, err error) {
	env, err := godotenv.Read(filepath.Join(dir, ".env"))
	if err != nil {
		return "", "", fmt.Errorf("failed to read .env file at %s: %w", filepath.Join(dir, ".env"), err)
	}
	apiPath, ok := env["API_PATH"]
	if !ok {
		return "", "", fmt.Errorf("environment variable API_PATH is not set")
	}

	name := path.Base(dir)

	ip, err := getIngressIP(name)
	if err != nil {
		common.PrintWarn("error getting ingress IP, falling back to local IP: %v", err)
		ip, err = common.GetLocalIP()
		if err != nil {
			return "", "", fmt.Errorf("error getting IP address: %w", err)
		}
	}

	gatewayURL, err = url.JoinPath(fmt.Sprintf("http://%s", ip), name, apiPath)
	if err != nil {
		return "", "", fmt.Errorf("error building gateway url: %w", err)
	}

	portalURL, err = url.JoinPath(fmt.Sprintf("http://%s", ip), name, "/dataportal/")
	if err != nil {
		return "", "", fmt.Errorf("error building dataportal url: %w", err)
	}

	return portalURL, gatewayURL, nil
}

func getIngressIP(namespace string) (string, error) {
	const (
		maxWait  = 30 * time.Second
		interval = 2 * time.Second
	)
	common.PrintStep("Waiting for ingress IP/hostname to be assigned...")
	start := time.Now()
	for {
		// Try to get the IP first
		cmd := exec.Command(
			"kubectl",
			"get", "ingress", "gateway",
			"-n", namespace,
			"-o", `jsonpath={.status.loadBalancer.ingress[0].ip}`,
		)
		out, err := common.RunCommand(cmd, true)
		if err != nil {
			return "", fmt.Errorf("error getting ingress ip: %w", err)
		}
		ip := strings.TrimSpace(out)
		if ip != "" {
			common.PrintDone("Ingress assigned IP: %s", ip)
			return ip, nil
		}

		// If IP is empty, try hostname
		cmd = exec.Command(
			"kubectl",
			"get", "ingress", "gateway",
			"-n", namespace,
			"-o", `jsonpath={.status.loadBalancer.ingress[0].hostname}`,
		)
		out, err = common.RunCommand(cmd, true)
		if err != nil {
			return "", fmt.Errorf("error getting ingress hostname: %w", err)
		}
		hostname := strings.TrimSpace(out)
		if hostname != "" {
			common.PrintDone("Ingress assigned hostname: %s", hostname)
			return hostname, nil
		}

		if time.Since(start) > maxWait {
			break
		}
		time.Sleep(interval)
	}
	return "", fmt.Errorf("error getting ingress IP: both ip and hostname are empty after waiting %s", maxWait)
}
