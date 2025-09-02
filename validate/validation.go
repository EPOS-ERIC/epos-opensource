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

var validHostnameRegex = regexp.MustCompile(
	`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`,
)

// enviroment name validation
func Name(name string) error {
	if err := validateEnvName(name); err != nil {
		return fmt.Errorf("invalid name for environment: %w", err)
	}
	return nil
}

// ip validation
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

// checks that an environment with the given name doesn't already exists otherwise returns an error
func EnviromentNotExistDocker(name string) error {
	_, err := db.GetDockerByName(name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("error getting installed docker environment from db: %w", err)
	}
	return fmt.Errorf("an enviroment with name '%s' already exists", name)
}

// checks that an environment with the given name exists
func EnviromentExistsDocker(name string) error {
	if err := EnviromentNotExistDocker(name); err != nil {
		return nil
	}
	return fmt.Errorf("no enviroment with name'%s' exists", name)
}

// checks that an environment with the given name doesn't already exists in kubernetes otherwise returns an error
func EnviromentNotExistK8s(name string) error {
	_, err := db.GetKubernetesByName(name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("error getting installed kubernetes environment from db: %w", err)
	}
	return fmt.Errorf("an enviroment with name '%s' already exists", name)
}

// path validation
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

var validEnv = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// ValidateEnvName returns an error if name is empty or contains anything
// other than alphanumerics, '.', '_' or '-'.
func validateEnvName(name string) error {
	if name == "" {
		return fmt.Errorf("environment name must not be empty")
	}

	if !validEnv.MatchString(name) {
		return fmt.Errorf("invalid environment name %s: only letters, digits, '.', '_' and '-' allowed", name)
	}
	return nil
}
