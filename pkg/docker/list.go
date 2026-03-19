package docker

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/db"
)

// List returns all persisted Docker environments.
func List() ([]Env, error) {
	rows, err := db.GetAllDocker()
	if err != nil {
		return nil, fmt.Errorf("failed to list docker environments: %w", err)
	}

	envs := make([]Env, 0, len(rows))
	for _, row := range rows {
		env, err := envFromDBRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to decode docker environment %q: %w", row.Name, err)
		}

		envs = append(envs, *env)
	}

	return envs, nil
}
