package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/db/sqlc"
)

// UpsertDocker adds a new docker entry to the database.
func UpsertDocker(docker sqlc.Docker) (*sqlc.Docker, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	d, err := q.UpsertDocker(context.Background(), sqlc.UpsertDockerParams(docker))
	if err != nil {
		return nil, fmt.Errorf("error inserting docker %s in db: %w", docker.Name, err)
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

// GetImageUpdateCache retrieves a cached remote image state for an image reference.
func GetImageUpdateCache(ctx context.Context, imageRef string) (*sqlc.ImageUpdateCache, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}

	cache, err := q.GetImageUpdateCache(ctx, imageRef)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting image update cache: %w", err)
	}

	return &cache, nil
}

// UpsertImageUpdateCache updates or inserts cached remote image state.
func UpsertImageUpdateCache(ctx context.Context, imageRef, remoteDigest string, remoteCreatedAt *time.Time, fetchedAt time.Time) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}

	err = q.UpsertImageUpdateCache(ctx, sqlc.UpsertImageUpdateCacheParams{
		ImageRef:        imageRef,
		RemoteDigest:    remoteDigest,
		RemoteCreatedAt: remoteCreatedAt,
		FetchedAt:       &fetchedAt,
	})
	if err != nil {
		return fmt.Errorf("error upserting image update cache: %w", err)
	}

	return nil
}

// InsertIngestedFile inserts or updates an ingested file record for a docker environment.
func InsertIngestedFile(envName, filePath string) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}
	err = q.InsertIngestedFile(context.Background(), sqlc.InsertIngestedFileParams{
		EnvironmentName: envName,
		FilePath:        filePath,
	})
	if err != nil {
		return fmt.Errorf("error inserting ingested file: %w", err)
	}
	return nil
}

// DeleteIngestedFilesByEnvironment deletes all ingested file records for an environment.
func DeleteIngestedFilesByEnvironment(envName string) error {
	q, err := Get()
	if err != nil {
		return fmt.Errorf("error getting db connection: %w", err)
	}
	err = q.DeleteIngestedFilesByEnvironment(context.Background(), envName)
	if err != nil {
		return fmt.Errorf("error deleting ingested files: %w", err)
	}
	return nil
}

// GetIngestedFilesByEnvironment retrieves all ingested file records for an environment.
func GetIngestedFilesByEnvironment(envName string) ([]sqlc.GetIngestedFilesByEnvironmentRow, error) {
	q, err := Get()
	if err != nil {
		return nil, fmt.Errorf("error getting db connection: %w", err)
	}
	files, err := q.GetIngestedFilesByEnvironment(context.Background(), envName)
	if err != nil {
		return nil, fmt.Errorf("error getting ingested files: %w", err)
	}
	return files, nil
}
