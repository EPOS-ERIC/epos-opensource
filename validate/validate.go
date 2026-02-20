// Package validate provides validation helpers for environment names and paths.
package validate

import (
	"fmt"
	"os"
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

// PathExists validates that a given path points to an existing directory on the filesystem.
// Returns an error if it doesn't.
func PathExists(path string) error {
	if path == "" {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %s does not exist", path)
		}
		return fmt.Errorf("cannot stat directory %s: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", path)
	}

	return nil
}
