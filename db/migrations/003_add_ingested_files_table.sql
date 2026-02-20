-- +goose Up
CREATE TABLE ingested_files (
    environment_type TEXT NOT NULL CHECK (environment_type IN ('docker', 'k8s')),
    environment_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (environment_type, environment_name, file_path)
);

-- +goose Down
DROP TABLE ingested_files;
