package docker

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/db"
)

// GetEnv fetches a deployed Docker environment by name.
func GetEnv(name string) (*Env, error) {
	row, err := db.GetDockerByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker environment %q: %w", name, err)
	}

	env, err := envFromDBRow(*row)
	if err != nil {
		return nil, fmt.Errorf("failed to decode docker environment %q: %w", name, err)
	}

	return env, nil
}
