// Package validate provides validation helpers for environment names and paths.
package validate

import (
	"fmt"
	"regexp"
)

var validEnv = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// Name validates environment name
func Name(name string) error {
	if name == "" {
		return fmt.Errorf("environment name must not be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("environment name %s is too long (max 63 characters)", name)
	}

	if !validEnv.MatchString(name) {
		return fmt.Errorf("invalid environment name %s: only lowercase letters, digits, and '-' allowed. Must fulfill regex %s", name, validEnv)
	}
	return nil
}
