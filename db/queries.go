package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/epos-eu/epos-opensource/db/sqlc"
)

// InsertKubernetes adds a new kubernetes entry to the database.
func InsertKubernetes(name, dir, contextStr, apiURL, guiURL, backofficeURL, protocol string) (*sqlc.Kubernetes, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	k, err := q.InsertKubernetes(context.Background(), sqlc.InsertKubernetesParams{
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
func GetKubernetesByName(name string) (*sqlc.Kubernetes, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	kube, err := q.GetKubernetesByName(context.Background(), name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error getting kubernetes %s: no row found: %w", name, err)
		}
		return nil, fmt.Errorf("error getting kubernetes %s: %w", name, err)
	}
	return &kube, nil
}

// GetAllKubernetes retrieves all kubernetes entries from the database.
func GetAllKubernetes() ([]sqlc.Kubernetes, error) {
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
func InsertDocker(docker sqlc.Docker) (*sqlc.Docker, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	d, err := q.InsertDocker(context.Background(), sqlc.InsertDockerParams{
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
func GetDockerByName(name string) (*sqlc.Docker, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	docker, err := q.GetDockerByName(context.Background(), name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error getting docker '%s' no row found: %w", name, err)
		}
		return nil, fmt.Errorf("error getting docker %s: %w", name, err)
	}
	return &docker, nil
}

// GetAllDocker retrieves all docker entries from the database.
func GetAllDocker() ([]sqlc.Docker, error) {
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

// GetLatestReleaseCache retrieves the latest release cache from the database.
func GetLatestReleaseCache() (sqlc.LatestReleaseCache, error) {
	q, err := Get()
	if err != nil {
		return sqlc.LatestReleaseCache{}, fmt.Errorf("error getting db connection: %w", err)
	}
	cache, err := q.GetLatestReleaseCache(context.Background())
	if err != nil {
		return sqlc.LatestReleaseCache{}, fmt.Errorf("error getting latest release cache: %w", err)
	}
	return cache, nil
}

// UpsertLatestReleaseCache updates or inserts the latest release cache in the database.
func UpsertLatestReleaseCache(tagName string, fetchedAt time.Time) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}
	err = q.UpsertLatestReleaseCache(context.Background(), sqlc.UpsertLatestReleaseCacheParams{
		TagName:   tagName,
		FetchedAt: &fetchedAt,
	})
	if err != nil {
		return fmt.Errorf("error upserting latest release cache: %w", err)
	}
	return nil
}

// InsertIngestedFile inserts or updates an ingested file record.
func InsertIngestedFile(envType, envName, filePath string) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}
	err = q.InsertIngestedFile(context.Background(), sqlc.InsertIngestedFileParams{
		EnvironmentType: envType,
		EnvironmentName: envName,
		FilePath:        filePath,
	})
	if err != nil {
		return fmt.Errorf("error inserting ingested file: %w", err)
	}
	return nil
}

// DeleteIngestedFilesByEnvironment deletes all ingested file records for an environment.
func DeleteIngestedFilesByEnvironment(envType, envName string) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}
	err = q.DeleteIngestedFilesByEnvironment(context.Background(), sqlc.DeleteIngestedFilesByEnvironmentParams{
		EnvironmentType: envType,
		EnvironmentName: envName,
	})
	if err != nil {
		return fmt.Errorf("error deleting ingested files: %w", err)
	}
	return nil
}

// GetIngestedFilesByEnvironment retrieves all ingested file records for an environment.
func GetIngestedFilesByEnvironment(envType, envName string) ([]sqlc.GetIngestedFilesByEnvironmentRow, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	files, err := q.GetIngestedFilesByEnvironment(context.Background(), sqlc.GetIngestedFilesByEnvironmentParams{
		EnvironmentType: envType,
		EnvironmentName: envName,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting ingested files: %w", err)
	}
	return files, nil
}
