package docker

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/db"
)

// EnsureEnvironmentDoesNotExist checks that the Docker environment does not exist.
func EnsureEnvironmentDoesNotExist(name string) error {
	_, err := db.GetDockerByName(name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("error getting installed docker environment from db: %w", err)
	}

	return fmt.Errorf("an environment with name '%s' already exists", name)
}

// EnsureEnvironmentExists checks that the Docker environment exists.
func EnsureEnvironmentExists(name string) error {
	_, err := db.GetDockerByName(name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no environment with name '%s' exists", name)
		}
		return fmt.Errorf("error getting installed docker environment from db: %w", err)
	}

	return nil
}
