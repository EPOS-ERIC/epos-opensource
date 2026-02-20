package k8s

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/command"
	"github.com/EPOS-ERIC/epos-opensource/common"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/EPOS-ERIC/epos-opensource/validate"
)

type CleanOpts struct {
	// Required. name of the environment
	Name string
	// Optional. Kubernetes context to use; defaults to the current kubectl context when unset.
	Context string
}

// TODO: clean up this function, remove weird stuff and reuse old functions instead if possible

func Clean(opts CleanOpts) (*Env, error) {
	if opts.Context == "" {
		context, err := common.GetCurrentKubeContext()
		if err != nil {
			return nil, fmt.Errorf("failed to get current kubectl context: %w", err)
		}
		opts.Context = context
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid clean parameters: %w", err)
	}

	display.Step("Cleaning data in environment: %s", opts.Name)

	env, err := GetEnv(opts.Name, opts.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment %s: %w", opts.Name, err)
	}

	cfg := env.Config
	cfg.Jobs.Enabled = true
	cfg.Jobs.InitDB.Enabled = true

	renderedFiles, err := cfg.Render()
	if err != nil {
		return nil, fmt.Errorf("failed to render chart templates: %w", err)
	}

	initDBManifest, err := getRenderedTemplateBySuffix(renderedFiles, "templates/init-db-job.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare init-db job manifest: %w", err)
	}

	ontologyManifest, err := getRenderedTemplateBySuffix(renderedFiles, "templates/ontology-populator-job.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare ontology populator job manifest: %w", err)
	}

	if err := runOneOffJobFromManifest(opts.Context, opts.Name, "init-db-job", initDBManifest, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to run init-db clean job: %w", err)
	}

	if err := runOneOffJobFromManifest(opts.Context, opts.Name, "ontology-populator-job", ontologyManifest, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to run ontology populator clean job: %w", err)
	}

	display.Done("Cleaned environment: %s", opts.Name)

	return env, nil
}

func (c *CleanOpts) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("environment name is required")
	}

	if err := validate.Name(c.Name); err != nil {
		return fmt.Errorf("'%s' is an invalid name for an environment: %w", c.Name, err)
	}

	if err := validate.EnvironmentExistsK8s(c.Name); err != nil {
		return fmt.Errorf("no environment with the name '%s' exists: %w", c.Name, err)
	}

	return nil
}

func getRenderedTemplateBySuffix(files map[string]string, suffix string) (string, error) {
	for name, content := range files {
		if !strings.HasSuffix(name, suffix) {
			continue
		}

		trimmed := strings.TrimSpace(content)
		if trimmed == "" || trimmed == "---" {
			return "", fmt.Errorf("template %s rendered empty output", name)
		}

		return content, nil
	}

	return "", fmt.Errorf("template with suffix %s not found in rendered output", suffix)
}

func runOneOffJobFromManifest(context, namespace, jobName, manifest string, timeout time.Duration) error {
	display.Step("Running clean job: %s", jobName)

	manifestPath, err := writeTempManifest(jobName, manifest)
	if err != nil {
		return fmt.Errorf("failed to write manifest for job %s: %w", jobName, err)
	}
	defer func() {
		if removeErr := os.Remove(manifestPath); removeErr != nil {
			display.Warn("failed to remove temporary manifest file %s: %v", manifestPath, removeErr)
		}
	}()

	if err := runKubectl(".", true, context, "delete", "job", jobName, "-n", namespace, "--ignore-not-found=true"); err != nil {
		return fmt.Errorf("failed to delete existing job %s before rerun: %w", jobName, err)
	}

	if err := runKubectl(".", false, context, "apply", "-f", manifestPath, "-n", namespace); err != nil {
		return fmt.Errorf("failed to apply job manifest for %s: %w", jobName, err)
	}

	timeoutSeconds := int(timeout.Seconds())
	if err := runKubectl(
		".",
		false,
		context,
		"wait",
		"--for=condition=complete",
		fmt.Sprintf("job/%s", jobName),
		"-n",
		namespace,
		"--timeout",
		fmt.Sprintf("%ds", timeoutSeconds),
	); err != nil {
		diagnostics := getJobDiagnostics(context, namespace, jobName)
		if diagnostics != "" {
			return fmt.Errorf("job %s did not complete successfully: %w\n%s", jobName, err, diagnostics)
		}
		return fmt.Errorf("job %s did not complete successfully: %w", jobName, err)
	}

	display.Done("Completed clean job: %s", jobName)

	return nil
}

func writeTempManifest(jobName, manifest string) (string, error) {
	file, err := os.CreateTemp("", fmt.Sprintf("epos-%s-*.yaml", jobName))
	if err != nil {
		return "", fmt.Errorf("failed to create temporary manifest file: %w", err)
	}

	if _, err := file.WriteString(manifest); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("failed to write temporary manifest file %s: %w", file.Name(), err)
	}

	if err := file.Close(); err != nil {
		return "", fmt.Errorf("failed to close temporary manifest file %s: %w", file.Name(), err)
	}

	return file.Name(), nil
}

func getJobDiagnostics(context, namespace, jobName string) string {
	describe := runKubectlCapture(context, "describe", "job", jobName, "-n", namespace)
	logs := runKubectlCapture(context, "logs", fmt.Sprintf("job/%s", jobName), "-n", namespace, "--all-containers=true")

	parts := []string{}
	if describe != "" {
		parts = append(parts, "kubectl describe output:\n"+describe)
	}
	if logs != "" {
		parts = append(parts, "kubectl logs output:\n"+logs)
	}

	return strings.Join(parts, "\n\n")
}

func runKubectlCapture(context string, args ...string) string {
	if context != "" {
		args = append(args, "--context", context)
	}

	out, err := command.RunCommand(exec.Command("kubectl", args...), true)
	if err != nil {
		trimmed := strings.TrimSpace(out)
		if trimmed != "" {
			return trimmed
		}
		return ""
	}

	return strings.TrimSpace(out)
}
