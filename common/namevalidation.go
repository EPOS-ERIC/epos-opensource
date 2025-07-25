package common

import (
	"fmt"
	"regexp"
)

// ValidateEnvName returns an error if name is empty or contains anything
// other than alphanumerics, '.', '_' or '-'.
func ValidateEnvName(name string) error {
	if name == "" {
		return fmt.Errorf("environment name must not be empty")
	}

	validEnv := regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	if !validEnv.MatchString(name) {
		return fmt.Errorf("invalid environment name %s: only letters, digits, '.', '_' and '-' allowed", name)
	}
	return nil
}
