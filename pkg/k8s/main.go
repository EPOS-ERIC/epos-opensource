package k8s

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/command"
	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/release"
)

// Env represents a deployed K8s environment and its effective configuration.
type Env struct {
	config.Config

	Name    string
	Context string
}

func runKubectl(dir string, suppressOut bool, context string, args ...string) error {
	cmd := newKubectlCommand(dir, context, args...)
	_, err := command.RunCommand(cmd, suppressOut)
	return err
}

func runKubectlWithInput(dir string, suppressOut bool, context, input string, args ...string) error {
	cmd := newKubectlCommand(dir, context, args...)
	cmd.Stdin = strings.NewReader(input)
	_, err := command.RunCommand(cmd, suppressOut)
	return err
}

func runKubectlCapture(dir, context string, args ...string) string {
	cmd := newKubectlCommand(dir, context, args...)
	out, err := command.RunCommand(cmd, true)
	if err != nil {
		trimmed := strings.TrimSpace(out)
		if trimmed != "" {
			return trimmed
		}
		return ""
	}

	return strings.TrimSpace(out)
}

func newKubectlCommand(dir, context string, args ...string) *exec.Cmd {
	if context != "" {
		args = append(args, "--context", context)
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Dir = dir
	return cmd
}

// deleteNamespace removes the specified namespace and all its resources.
func deleteNamespace(name, context string) error {
	if err := runKubectl(".", false, context, "delete", "namespace", name); err != nil {
		return fmt.Errorf("failed to delete namespace %s: %w", name, err)
	}
	return nil
}

// ConfigFromRelease converts a Helm release values map into a typed K8s Config.
func ConfigFromRelease(rel *release.Release) (*config.Config, error) {
	if rel == nil {
		return nil, fmt.Errorf("release is nil")
	}

	yamlBytes, err := yaml.Marshal(rel.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config to yaml: %w", err)
	}

	var envConfig config.Config
	if err := yaml.Unmarshal(yamlBytes, &envConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &envConfig, nil
}

// EnvFromRelease converts a Helm release to an Env bound to the provided kube context.
func EnvFromRelease(rel *release.Release, context string) (*Env, error) {
	if rel == nil {
		return nil, fmt.Errorf("release is nil")
	}

	envConfig, err := ConfigFromRelease(rel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert release to config: %w", err)
	}

	return &Env{
		Config:  *envConfig,
		Name:    envConfig.Name,
		Context: context,
	}, nil
}
