-- Docker queries
-- name: GetAllDocker :many
SELECT
    name,
    config_yaml
FROM
    docker
ORDER BY
    name;

-- name: UpsertDocker :one
INSERT INTO
    docker (
        name,
        config_yaml
    )
VALUES
    (?, ?) ON CONFLICT (name) DO
UPDATE
SET
    config_yaml = excluded.config_yaml
RETURNING
    name,
    config_yaml;

-- name: DeleteDocker :exec
DELETE FROM
    docker
WHERE
    name = ?;

-- name: GetDockerByName :one
SELECT
    name,
    config_yaml
FROM
    docker
WHERE
    name = ?;

-- name: GetLatestReleaseCache :one
SELECT
    *
FROM
    latest_release_cache
WHERE
    id = 1;

-- name: UpsertLatestReleaseCache :exec
INSERT INTO
    latest_release_cache (id, tag_name, fetched_at)
VALUES
    (1, ?, ?) ON CONFLICT (id) DO
UPDATE
SET
    tag_name = excluded.tag_name,
    fetched_at = excluded.fetched_at;

-- name: InsertIngestedFile :exec
INSERT INTO
    ingested_files (
        environment_name,
        file_path,
        ingested_at
    )
VALUES
    (?, ?, CURRENT_TIMESTAMP) ON CONFLICT (environment_name, file_path) DO
UPDATE
SET
    ingested_at = CURRENT_TIMESTAMP;

-- name: DeleteIngestedFilesByEnvironment :exec
DELETE FROM
    ingested_files
WHERE
    environment_name = ?;

-- name: GetIngestedFilesByEnvironment :many
SELECT
    file_path,
    ingested_at
FROM
    ingested_files
WHERE
    environment_name = ?
ORDER BY
    ingested_at DESC;
