package validate

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"

	"github.com/epos-eu/epos-opensource/db"
)

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
func CustomHost(CustomHost string) error {
	if CustomHost == "" {
		return nil
	}

	ip := net.ParseIP(CustomHost)
	if ip == nil {
		if !validHostnameRegex.MatchString(CustomHost) {
			return fmt.Errorf("custom host %q is not a valid ip or hostname", CustomHost)
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

// EnvironmentNotExistK8s checks that Kubernetes environment doesn't exist
func EnvironmentNotExistK8s(name string) error {
	_, err := db.GetKubernetesByName(name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("error getting installed kubernetes environment from db: %w", err)
	}
	return fmt.Errorf("an environment with name '%s' already exists", name)
}

// EnvironmentExistsK8s checks that Kubernetes environment exists
func EnvironmentExistsK8s(name string) error {
	if err := EnvironmentNotExistK8s(name); err != nil {
		return nil
	}
	return fmt.Errorf("no environment with name '%s' exists", name)
}

// PathExists validates that path is an existing directory
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
