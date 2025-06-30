package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path"

	"github.com/epos-eu/epos-opensource/common/configdir"
	_ "modernc.org/sqlite"
)

//go:generate go tool sqlc generate -f ../sqlc.yaml

//go:embed schema.sql
var schema string

const dbName = "db.db"

// Get opens a new connection to the database, creating the database file and schema if they do not exist.
func Get() (*Queries, error) {
	dbDir := configdir.GetConfigPath()
	dbFile := path.Join(dbDir, dbName)

	err := os.MkdirAll(dbDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating db directory %s: %w", dbDir, err)
	}

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		file, err := os.Create(dbFile)
		if err != nil {
			return nil, fmt.Errorf("error creating db file %s: %w", dbFile, err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("error opening sqlite db %s: %w", dbFile, err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("error creating db schema: %w", err)
	}

	queries := New(db)

	return queries, nil
}

// DeleteEnv removes an environment entry from the database for the given name and platform.
func DeleteEnv(name, platform string) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}

	err = q.DeleteEnv(context.Background(), DeleteEnvParams{
		Name:     name,
		Platform: platform,
	})
	if err != nil {
		return fmt.Errorf("error deleting env %s for platform %s from db: %w", name, platform, err)
	}

	return nil
}

// InsertEnv adds a new environment entry to the database.
func InsertEnv(name, dir, platform string) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}

	_, err = q.InsertEnv(context.Background(), InsertEnvParams{
		Name:      name,
		Directory: dir,
		Platform:  platform,
	})
	if err != nil {
		return fmt.Errorf("error inserting env %s (dir: %s, platform: %s) in db: %w", name, dir, platform, err)
	}

	return nil
}

// GetEnvs retrieves all environment entries for the specified platform from the database.
func GetEnvs(platform string) ([]Environment, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}

	envs, err := q.GetPlatformEnvs(context.Background(), platform)
	if err != nil {
		return nil, fmt.Errorf("error getting platform envs for %s: %w", platform, err)
	}

	return envs, nil
}

// GetEnv retrieves a single environment entry for the specified name and platform from the database.
func GetEnv(name, platform string) (*Environment, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	env, err := q.GetEnvByNameAndPlatform(context.Background(), GetEnvByNameAndPlatformParams{
		Name:     name,
		Platform: platform,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting env %s for platform %s: %w", name, platform, err)
	}
	return &env, nil
}
