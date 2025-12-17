-- +goose Up
CREATE TABLE ingested_files (
    environment_type TEXT NOT NULL CHECK (environment_type IN ('docker', 'kubernetes')),
    environment_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (environment_type, environment_name, file_path),
    FOREIGN KEY (environment_name) REFERENCES docker(name) ON DELETE CASCADE,
    FOREIGN KEY (environment_name) REFERENCES kubernetes(name) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE ingested_files;

