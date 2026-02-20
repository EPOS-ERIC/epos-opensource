package k8s

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/command"
)

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
