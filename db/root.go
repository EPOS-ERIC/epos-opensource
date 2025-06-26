package db

import (
	"context"
	"database/sql"
	_ "embed"
	"os"
	"path"

	_ "modernc.org/sqlite"
)

//go:generate go tool sqlc generate -f ../sqlc.yaml

//go:embed schema.sql
var schema string

const (
	dir    = "~/.epos-opensource"
	dbName = "db.db"
)

// Get opens a new connection to the db. if it does not exist in the file system it is created.
func Get() (*Queries, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dbDir := path.Join(homeDir, ".epos-opensource")
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		return nil, err
	}
	dbFile := path.Join(dbDir, dbName)

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		file, err := os.Create(dbFile)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	queries := New(db)

	return queries, nil
}

func DeleteEnv(name, platform string) error {
	q, err := Get()
	if err != nil {
		return err
	}

	err = q.DeleteEnv(context.Background(), DeleteEnvParams{
		Name:     name,
		Platform: "docker",
	})
	if err != nil {
		return err
	}

	return nil
}

func InsertEnv(name, dir, platform string) error {
	q, err := Get()
	if err != nil {
		return err
	}

	_, err = q.InsertEnv(context.Background(), InsertEnvParams{
		Name:      name,
		Directory: dir,
		Platform:  platform,
	})
	if err != nil {
		return err
	}

	return nil
}

func GetEnvs(platform string) ([]Environment, error) {
	q, err := Get()
	if err != nil {
		return nil, err
	}

	envs, err := q.GetPlatformEnvs(context.Background(), platform)
	if err != nil {
		return nil, err
	}

	return envs, nil
}
