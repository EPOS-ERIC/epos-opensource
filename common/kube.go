package common

import (
	"os/exec"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/command"
)

// GetCurrentKubeContext retrieves the current kubectl context.
func GetCurrentKubeContext() (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")
	out, err := command.RunCommand(cmd, true)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetKubeContexts retrieves the list of available kubectl contexts.
func GetKubeContexts() ([]string, error) {
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o=name")
	out, err := command.RunCommand(cmd, true)
	if err != nil {
		return nil, err
	}
	contexts := strings.Split(strings.TrimSpace(string(out)), "\n")
	return contexts, nil
}
