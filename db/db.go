// Package db manages the SQLite database for epos-opensource.
// It embeds the SQL schema, ensures the database file is created under the configured data directory,
// and provides functions to open the connection and perform CRUD operations on Kubernetes and Docker entries.
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path"

	"github.com/epos-eu/epos-opensource/config"
	"github.com/epos-eu/epos-opensource/db/sqlc"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:generate go tool sqlc generate -f ../sqlc.yaml

//go:embed migrations/*.sql
var embedMigrations embed.FS

const dbName = "db.db"

func init() {
	goose.SetBaseFS(embedMigrations)
	err := goose.SetDialect("sqlite")
	if err != nil {
		panic("if you see this error please open an issue on github: " + err.Error())
	}
}

// Get opens a new connection to the database, creating the database file and schema if they do not exist.
func Get() (*sqlc.Queries, error) {
	dbDir := config.GetDataPath()
	dbFile := path.Join(dbDir, dbName)

	err := os.MkdirAll(dbDir, 0o750)
	if err != nil {
		return nil, fmt.Errorf("error creating db directory %s: %w", dbDir, err)
	}

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		file, err := os.Create(dbFile)
		if err != nil {
			return nil, fmt.Errorf("error creating db file %s: %w", dbFile, err)
		}
		err = file.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close db file: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("error opening sqlite db %s: %w", dbFile, err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		newErr := db.Close()
		if newErr != nil {
			return nil, fmt.Errorf("failed to close db on ping error '%w': %w", err, newErr)
		}
		return nil, fmt.Errorf("failed to ping database %s: %w", dbFile, err)
	}

	if err := runMigrations(db); err != nil {
		newErr := db.Close()
		if newErr != nil {
			return nil, fmt.Errorf("failed to close db on migrations error '%w': %w", err, newErr)
		}
		return nil, err
	}

	queries := sqlc.New(db)
	return queries, nil
}

func runMigrations(db *sql.DB) error {
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to execute database migrations: %w", err)
	}
	return nil
}
