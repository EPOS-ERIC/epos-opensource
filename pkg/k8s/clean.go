package k8s

import (
	"fmt"
	"strings"
	"time"

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

func Clean(opts CleanOpts) (*Env, error) {
	if opts.Context == "" {
		context, err := common.GetCurrentKubeContext()
		if err != nil {
			return nil, fmt.Errorf("failed to get current kubectl context: %w", err)
		}

		opts.Context = context

		display.Debug("using current kubectl context: %s", opts.Context)
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid clean parameters: %w", err)
	}

	display.Step("Cleaning data in environment: %s", opts.Name)

	env, err := GetEnv(opts.Name, opts.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment %s: %w", opts.Name, err)
	}

	display.Debug("loaded environment: %s", env.Name)

	cfg := env.Config
	cfg.Jobs.Enabled = true
	cfg.Jobs.InitDB.Enabled = true

	display.Debug("enabled init-db and job execution for clean")
	display.Debug("rendering chart templates for clean jobs")

	renderedFiles, err := cfg.Render()
	if err != nil {
		return nil, fmt.Errorf("failed to render chart templates: %w", err)
	}

	display.Debug("rendered templates: %d", len(renderedFiles))

	jobs := []struct {
		name       string
		template   string
		errorLabel string
	}{
		{name: "init-db-job", template: "templates/init-db-job.yaml", errorLabel: "init-db"},
		{name: "ontology-populator-job", template: "templates/ontology-populator-job.yaml", errorLabel: "ontology populator"},
	}

	for _, job := range jobs {
		display.Debug("preparing clean job manifest: %s", job.template)

		manifest, err := getRenderedTemplateBySuffix(renderedFiles, job.template)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare %s job manifest: %w", job.errorLabel, err)
		}

		display.Debug("running clean job: %s", job.name)

		if err := runOneOffJobFromManifest(opts.Context, opts.Name, job.name, manifest, 5*time.Minute); err != nil {
			return nil, fmt.Errorf("failed to run %s clean job: %w", job.errorLabel, err)
		}
	}

	display.Done("Cleaned environment: %s", opts.Name)

	return env, nil
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

	if err := runKubectl(".", true, context, "delete", "job", jobName, "-n", namespace, "--ignore-not-found=true"); err != nil {
		return fmt.Errorf("failed to delete existing job %s before rerun: %w", jobName, err)
	}

	if err := runKubectlWithInput(".", false, context, manifest, "apply", "-f", "-", "-n", namespace); err != nil {
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
		describe := runKubectlCapture(".", context, "describe", "job", jobName, "-n", namespace)
		logs := runKubectlCapture(".", context, "logs", fmt.Sprintf("job/%s", jobName), "-n", namespace, "--all-containers=true")

		parts := []string{}
		if describe != "" {
			parts = append(parts, "kubectl describe output:\n"+describe)
		}
		if logs != "" {
			parts = append(parts, "kubectl logs output:\n"+logs)
		}

		diagnostics := strings.Join(parts, "\n\n")
		if diagnostics != "" {
			return fmt.Errorf("job %s did not complete successfully: %w\n%s", jobName, err, diagnostics)
		}
		return fmt.Errorf("job %s did not complete successfully: %w", jobName, err)
	}

	display.Done("Completed clean job: %s", jobName)

	return nil
}

func (c *CleanOpts) Validate() error {
	display.Debug("name: %s", c.Name)
	display.Debug("context: %s", c.Context)

	if c.Name == "" {
		return fmt.Errorf("environment name is required")
	}

	if err := validate.Name(c.Name); err != nil {
		return fmt.Errorf("'%s' is an invalid name for an environment: %w", c.Name, err)
	}

	if c.Context == "" {
		context, err := common.GetCurrentKubeContext()
		if err != nil {
			return fmt.Errorf("failed to get current kubectl context: %w", err)
		}

		c.Context = context

		display.Debug("using current kubectl context: %s", c.Context)
	} else if err := EnsureContextExists(c.Context); err != nil {
		return fmt.Errorf("K8s context %q is not an available context: %w", c.Context, err)
	}

	if err := EnsureEnvironmentExists(c.Name, c.Context); err != nil {
		return fmt.Errorf("no environment with the name '%s' exists: %w", c.Name, err)
	}

	return nil
}
