package internal

import (
	"embed"
	"epos-cli/common"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"
)

const pathPrefix = "k8s"

// Path inside the embed.FS where the manifests live.
const embedManifestsPath = "static/manifests"

//go:embed static/manifests
var manifestsFs embed.FS

// EmbeddedManifestContents holds the raw text of each embedded manifest file keyed by filename.
var EmbeddedManifestContents map[string]string

//go:embed static/.env
var envFile string

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

		// Store the manifest content keyed by the original filename.
		EmbeddedManifestContents[entry.Name()] = string(data)
	}
}

// NewEnvDir creates a new environment directory with .env and manifest files.
// If customEnvFilePath is provided, it reads the .env content from that file path,
// otherwise uses the embedded default .env content.
// If customManifestsDirPath is provided, it copies all files from that directory,
// otherwise uses the embedded default manifest files.
// Returns the path to the created environment directory.
// If any error occurs after directory creation, the directory and its contents are automatically cleaned up.
func NewEnvDir(customEnvFilePath, customManifestsDirPath, customPath, name string) (string, error) {
	envPath := common.BuildEnvPath(customPath, name, pathPrefix)

	// Check if directory already exists
	if _, err := os.Stat(envPath); err == nil {
		return "", fmt.Errorf("directory %s already exists", envPath)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check directory %s: %w", envPath, err)
	}

	if err := os.MkdirAll(envPath, 0700); err != nil {
		return "", fmt.Errorf("failed to create env directory %s: %w", envPath, err)
	}

	var ok bool
	// Ensure cleanup of directory if any error occurs after creation
	defer func() {
		if !ok {
			if removeErr := os.RemoveAll(envPath); removeErr != nil {
				common.PrintError("Failed to cleanup directory '%s' after error: %v. You may need to remove it manually.", envPath, removeErr)
			}
		}
	}()

	envContent, err := common.GetContentFromPathOrDefault(customEnvFilePath, envFile)
	if err != nil {
		return "", fmt.Errorf("failed to get .env file content: %w", err)
	}

	if err := common.CreateFileWithContent(path.Join(envPath, ".env"), envContent); err != nil {
		return "", fmt.Errorf("failed to create .env file: %w", err)
	}

	if customManifestsDirPath != "" {
		dirEntries, err := os.ReadDir(customManifestsDirPath)
		if err != nil {
			return "", fmt.Errorf("error reading manifests dir: %w", err)
		}

		for _, de := range dirEntries {
			if de.IsDir() {
				continue
			}

			srcPath := filepath.Join(customManifestsDirPath, de.Name())
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return "", fmt.Errorf("error reading file %q: %w", de.Name(), err)
			}
			if err := common.CreateFileWithContent(filepath.Join(envPath, de.Name()), string(data)); err != nil {
				return "", fmt.Errorf("failed to create manifest file: %w", err)
			}
		}
	} else {
		for filename, content := range EmbeddedManifestContents {
			if err := common.CreateFileWithContent(filepath.Join(envPath, filename), content); err != nil {
				return "", fmt.Errorf("failed to create embedded manifest %q: %w", filename, err)
			}
		}
	}

	ok = true
	return envPath, nil
}

// runKubectl executes `kubectl ...` inside the given dir with the supplied context
func runKubectl(dir string, args ...string) error {
	cmd := exec.Command("kubectl", args...)
	cmd.Dir = dir
	return common.RunCommand(cmd)
}

// applyParallel shells out to `kubectl apply` for every element in targets.
// If withService == true it expects YAML pairs (deployment‑X.yaml/service‑X.yaml)
// for each target name; otherwise it applies the file supplied.
func applyParallel(dir string, targets []string, withService bool) error {
	var g errgroup.Group

	for _, t := range targets {
		t := t // capture
		g.Go(func() error {
			if withService {
				return runKubectl(dir,
					"apply", "-f", fmt.Sprintf("deployment-%s.yaml", t), "-f", fmt.Sprintf("service-%s.yaml", t))
			}
			return runKubectl(dir, "apply", "-f", t)
		})
	}
	return g.Wait()
}

// waitDeployments blocks until every deployment in names reports a successful
// rollout (like `kubectl rollout status …`). It mirrors the parallel wait logic
// of the Bash script using a goroutine per deployment.
func waitDeployments(dir, namespace string, names []string) error {
	var g errgroup.Group

	for _, n := range names {
		n := n
		g.Go(func() error {
			return runKubectl(dir, "rollout", "status", fmt.Sprintf("deployment/%s", n), "--timeout", (1 * time.Minute).String(), "-n", namespace)
		})
	}
	return g.Wait()
}

// TODO: drop the kubectl dependency and replace the shell execs with k8s/client‑go
func deployStack(dir, namespace string) error {
	setup := []string{
		"namespace.yaml",
		"configmap-epos-env.yaml",
		"secret-epos-secret.yaml",
		"pvc-psqldata.yaml",
		"pvc-converter-plugins.yaml",
	}
	for _, f := range setup {
		if err := runKubectl(dir, "apply", "-f", f); err != nil {
			return fmt.Errorf("apply %s: %w", f, err)
		}
	}

	infra := []string{"rabbitmq", "metadata-database"}
	common.PrintStep("Deploying infrastructure components")
	if err := applyParallel(dir, infra, true); err != nil {
		return err
	}

	if err := runKubectl(dir, "apply", "-f", "service-rabbitmq-management.yaml"); err != nil {
		return err
	}

	common.PrintStep("Waiting for infrastructure to be ready")
	if err := waitDeployments(dir, namespace, infra); err != nil {
		return err
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
		return err
	}

	common.PrintStep("Waiting for services to be ready")
	if err := waitDeployments(dir, namespace, services); err != nil {
		return err
	}

	finals := []string{"gateway", "dataportal"}
	common.PrintStep("Deploying gateway and dataportal")
	if err := applyParallel(dir, finals, true); err != nil {
		return err
	}
	if err := waitDeployments(dir, namespace, finals); err != nil {
		return err
	}

	common.PrintDone("EPOS platform deployed")
	return nil
}
