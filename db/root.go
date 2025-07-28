// Package db manages the SQLite database for epos-opensource.
// It embeds the SQL schema, ensures the database file is created under the configured data directory,
// and provides functions to open the connection and perform CRUD operations on Kubernetes and Docker entries.
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
	dbDir := configdir.GetPath()
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

// InsertKubernetes adds a new kubernetes entry to the database.
func InsertKubernetes(name, dir, contextStr, apiURL, guiURL, backofficeURL, protocol string) (*Kubernetes, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	k, err := q.InsertKubernetes(context.Background(), InsertKubernetesParams{
		Name:          name,
		Directory:     dir,
		Context:       contextStr,
		ApiUrl:        apiURL,
		GuiUrl:        guiURL,
		Protocol:      protocol,
		BackofficeUrl: backofficeURL,
	})
	if err != nil {
		return nil, fmt.Errorf("error inserting kubernetes %s (dir: %s) in db: %w", name, dir, err)
	}
	return &k, nil
}

// DeleteKubernetes removes a kubernetes entry from the database for the given name.
func DeleteKubernetes(name string) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}
	err = q.DeleteKubernetes(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error deleting kubernetes %s from db: %w", name, err)
	}
	return nil
}

// GetKubernetesByName retrieves a single kubernetes entry by name from the database.
func GetKubernetesByName(name string) (*Kubernetes, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	kube, err := q.GetKubernetesByName(context.Background(), name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("error getting kubernetes %s: no row found", name)
		}
		return nil, fmt.Errorf("error getting kubernetes %s: %w", name, err)
	}
	return &kube, nil
}

// GetAllKubernetes retrieves all kubernetes entries from the database.
func GetAllKubernetes() ([]Kubernetes, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	kubes, err := q.GetAllKubernetes(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting all kubernetes: %w", err)
	}
	return kubes, nil
}

// InsertDocker adds a new docker entry to the database.
func InsertDocker(docker Docker) (*Docker, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	d, err := q.InsertDocker(context.Background(), InsertDockerParams{
		Name:           docker.Name,
		Directory:      docker.Directory,
		ApiUrl:         docker.ApiUrl,
		GuiUrl:         docker.GuiUrl,
		BackofficeUrl:  docker.BackofficeUrl,
		GuiPort:        docker.GuiPort,
		ApiPort:        docker.ApiPort,
		BackofficePort: docker.BackofficePort,
	})
	if err != nil {
		return nil, fmt.Errorf("error inserting docker %s (dir: %s) in db: %w", docker.Name, docker.Directory, err)
	}
	return &d, nil
}

// DeleteDocker removes a docker entry from the database for the given name.
func DeleteDocker(name string) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}
	err = q.DeleteDocker(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error deleting docker %s from db: %w", name, err)
	}
	return nil
}

// GetDockerByName retrieves a single docker entry by name from the database.
func GetDockerByName(name string) (*Docker, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	docker, err := q.GetDockerByName(context.Background(), name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("error getting docker %s: no row found", name)
		}
		return nil, fmt.Errorf("error getting docker %s: %w", name, err)
	}
	return &docker, nil
}

// GetAllDocker retrieves all docker entries from the database.
func GetAllDocker() ([]Docker, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	dockers, err := q.GetAllDocker(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting all docker: %w", err)
	}
	return dockers, nil
}

func UpdateDocker(docker Docker) (*Docker, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}

	result, err := q.UpdateDocker(context.Background(), UpdateDockerParams{
		Directory:      docker.Directory,
		ApiUrl:         docker.ApiUrl,
		GuiUrl:         docker.GuiUrl,
		BackofficeUrl:  docker.BackofficeUrl,
		ApiPort:        docker.ApiPort,
		GuiPort:        docker.GuiPort,
		BackofficePort: docker.BackofficePort,
		Name:           docker.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting all docker: %w", err)
	}
	return &result, nil
}
