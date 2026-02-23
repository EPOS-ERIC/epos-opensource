package docker

import (
	"fmt"

	"github.com/EPOS-ERIC/epos-opensource/db"
	"github.com/EPOS-ERIC/epos-opensource/db/sqlc"
	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/config"
)

// Env represents a deployed Docker environment and its effective configuration.
type Env struct {
	config.EnvConfig

	Name string
}

func envFromDBRow(row sqlc.Docker) (*Env, error) {
	cfg, err := config.LoadConfigFromBytes([]byte(row.ConfigYaml))
	if err != nil {
		return nil, fmt.Errorf("failed to parse stored docker config: %w", err)
	}

	if cfg.Name == "" {
		cfg.Name = row.Name
	}

	if cfg.Name != row.Name {
		return nil, fmt.Errorf("stored config name %q does not match environment name %q", cfg.Name, row.Name)
	}

	return &Env{
		EnvConfig: *cfg,
		Name:      row.Name,
	}, nil
}

func upsertEnvConfig(cfg *config.EnvConfig) (*Env, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	if cfg.Name == "" {
		return nil, fmt.Errorf("environment name is required")
	}

	bytes, err := cfg.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	stored, err := db.UpsertDocker(sqlc.Docker{
		Name:       cfg.Name,
		ConfigYaml: string(bytes),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to persist environment config: %w", err)
	}

	env, err := envFromDBRow(*stored)
	if err != nil {
		return nil, fmt.Errorf("failed to decode persisted environment config: %w", err)
	}

	return env, nil
}
