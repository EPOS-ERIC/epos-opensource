// Package validate provides validation helpers for environment names and paths.
package validate

import (
	"fmt"
	"regexp"
)

var validEnv = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// Name validates environment name
func Name(name string) error {
	if name == "" {
		return fmt.Errorf("environment name must not be empty")
	}

	if !validEnv.MatchString(name) {
		return fmt.Errorf("invalid environment name %s: only letters, digits, '.', '_' and '-' allowed", name)
	}
	return nil
}
