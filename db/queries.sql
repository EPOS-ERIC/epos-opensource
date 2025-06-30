-- name: GetAllEnvs :many
SELECT
    *
FROM
    environment;

-- name: InsertEnv :one
INSERT INTO
    environment (name, directory, platform)
VALUES
    (?, ?, ?)
RETURNING
    *;

-- name: DeleteEnv :exec
DELETE FROM
    environment
WHERE
    name = ?
    AND platform = ?;

-- name: GetPlatformEnvs :many
SELECT
    *
FROM
    environment
WHERE
    platform = ?;

-- name: GetEnvByNameAndPlatform :one
SELECT
    *
FROM
    environment
WHERE
    name = ?
    AND platform = ?;

