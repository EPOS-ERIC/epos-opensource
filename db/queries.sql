-- K8s queries
-- name: GetAllK8s :many
SELECT
    *
FROM
    k8s;

-- name: InsertK8s :one
INSERT INTO
    k8s (
        name,
        directory,
        context,
        api_url,
        gui_url,
        backoffice_url,
        protocol,
        tls_enabled
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING
    *;

-- name: DeleteK8s :exec
DELETE FROM
    k8s
WHERE
    name = ?;

-- name: GetK8sByName :one
SELECT
    *
FROM
    k8s
WHERE
    name = ?;

-- Docker queries
-- name: GetAllDocker :many
SELECT
    *
FROM
    docker;

-- name: UpsertDocker :one
INSERT INTO
    docker (
        name,
        directory,
        api_url,
        gui_url,
        backoffice_url,
        gui_port,
        api_port,
        backoffice_port
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT (name) DO
UPDATE
SET
    directory = excluded.directory,
    api_url = excluded.api_url,
    gui_url = excluded.gui_url,
    backoffice_url = excluded.backoffice_url,
    gui_port = excluded.gui_port,
    api_port = excluded.api_port,
    backoffice_port = excluded.backoffice_port
RETURNING
    *;

-- name: DeleteDocker :exec
DELETE FROM
    docker
WHERE
    name = ?;

-- name: GetDockerByName :one
SELECT
    *
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
        environment_type,
        environment_name,
        file_path,
        ingested_at
    )
VALUES
    (?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT (environment_type, environment_name, file_path) DO
UPDATE
SET
    ingested_at = CURRENT_TIMESTAMP;

-- name: DeleteIngestedFilesByEnvironment :exec
DELETE FROM
    ingested_files
WHERE
    environment_type = ?
    AND environment_name = ?;

-- name: GetIngestedFilesByEnvironment :many
SELECT
    file_path,
    ingested_at
FROM
    ingested_files
WHERE
    environment_type = ?
    AND environment_name = ?
ORDER BY
    ingested_at DESC;
