package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/epos-eu/epos-opensource/db/sqlc"
)

// InsertK8s adds a new k8s entry to the database.
func InsertK8s(name, dir, contextStr, apiURL, guiURL, backofficeURL, protocol string) (*sqlc.K8s, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	k, err := q.InsertK8s(context.Background(), sqlc.InsertK8sParams{
		Name:          name,
		Directory:     dir,
		Context:       contextStr,
		ApiUrl:        apiURL,
		GuiUrl:        guiURL,
		Protocol:      protocol,
		BackofficeUrl: backofficeURL,
	})
	if err != nil {
		return nil, fmt.Errorf("error inserting k8s %s (dir: %s) in db: %w", name, dir, err)
	}
	return &k, nil
}

// DeleteK8s removes a k8s entry from the database for the given name.
func DeleteK8s(name string) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}
	err = q.DeleteK8s(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error deleting k8s %s from db: %w", name, err)
	}
	return nil
}

// GetK8sByName retrieves a single k8s entry by name from the database.
func GetK8sByName(name string) (*sqlc.K8s, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	k, err := q.GetK8sByName(context.Background(), name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error getting k8s %s: no row found: %w", name, err)
		}
		return nil, fmt.Errorf("error getting k8s %s: %w", name, err)
	}
	return &k, nil
}

// GetAllK8s retrieves all k8s entries from the database.
func GetAllK8s() ([]sqlc.K8s, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	ks, err := q.GetAllK8s(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting all k8s: %w", err)
	}
	return ks, nil
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
