package k8s

import (
	"fmt"
	"os/exec"

	"github.com/EPOS-ERIC/epos-opensource/command"

	"golang.org/x/sync/errgroup"
)

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

// deleteNamespace removes the specified namespace and all its resources.
func deleteNamespace(name, context string) error {
	if err := runKubectl(".", false, context, "delete", "namespace", name); err != nil {
		return fmt.Errorf("failed to delete namespace %s: %w", name, err)
	}
	return nil
}
