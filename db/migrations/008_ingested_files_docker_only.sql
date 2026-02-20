-- +goose Up
CREATE TABLE ingested_files_new (
    environment_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (environment_name, file_path)
);

INSERT INTO
    ingested_files_new (environment_name, file_path, ingested_at)
SELECT
    environment_name,
    file_path,
    ingested_at
FROM
    ingested_files
WHERE
    environment_type = 'docker';

DROP TABLE ingested_files;

ALTER TABLE
    ingested_files_new RENAME TO ingested_files;

-- +goose Down
CREATE TABLE ingested_files_new (
    environment_type TEXT NOT NULL CHECK (environment_type IN ('docker', 'k8s')),
    environment_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (environment_type, environment_name, file_path)
);

INSERT INTO
    ingested_files_new (environment_type, environment_name, file_path, ingested_at)
SELECT
    'docker',
    environment_name,
    file_path,
    ingested_at
FROM
    ingested_files;

DROP TABLE ingested_files;

ALTER TABLE
    ingested_files_new RENAME TO ingested_files;
