package validate

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"

	"github.com/EPOS-ERIC/epos-opensource/db"
)

// TODO: we have to update the validation considering the new refactored version, for example k8s won't have the db anymore

var (
	validHostnameRegex = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
	validEnv           = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
)

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

// CustomHost validates IP address or hostname
func CustomHost(customHost string) error {
	if customHost == "" {
		return nil
	}

	ip := net.ParseIP(customHost)
	if ip == nil {
		if !validHostnameRegex.MatchString(customHost) {
			return fmt.Errorf("custom host %q is not a valid ip or hostname", customHost)
		}
	}

	return nil
}

// EnvironmentNotExistDocker checks that Docker environment doesn't exist
func EnvironmentNotExistDocker(name string) error {
	_, err := db.GetDockerByName(name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("error getting installed docker environment from db: %w", err)
	}
	return fmt.Errorf("an environment with name '%s' already exists", name)
}

// EnvironmentExistsDocker checks that Docker environment exists
func EnvironmentExistsDocker(name string) error {
	if err := EnvironmentNotExistDocker(name); err != nil {
		return nil
	}
	return fmt.Errorf("no environment with name'%s' exists", name)
}

// EnvironmentNotExistK8s checks that K8s environment doesn't exist
func EnvironmentNotExistK8s(name string) error {
	// TODO
	// _, err := db.GetK8sByName(name)
	// if err != nil {
	// 	if errors.Is(err, sql.ErrNoRows) {
	// 		return nil
	// 	}
	// 	return fmt.Errorf("error getting installed k8s environment from db: %w", err)
	// }
	// return fmt.Errorf("an environment with name '%s' already exists", name)
	return nil
}

// EnvironmentExistsK8s checks that K8s environment exists
func EnvironmentExistsK8s(name string) error {
	// TODO
	// if err := EnvironmentNotExistK8s(name); err != nil {
	// 	return nil
	// }
	// return fmt.Errorf("no environment with name '%s' exists", name)
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

// IsFile validates that path is an existing file
func IsFile(path string) error {
	if path == "" {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", path)
		}
		return fmt.Errorf("failed to stat path %q: %w", path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("path %q is a directory, not a file", path)
	}

	return nil
}
