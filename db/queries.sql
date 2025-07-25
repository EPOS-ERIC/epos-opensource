-- Kubernetes queries
-- name: GetAllKubernetes :many
SELECT
    *
FROM
    kubernetes;

-- name: InsertKubernetes :one
INSERT INTO
    kubernetes (
        name,
        directory,
        context,
        api_url,
        gui_url,
        backoffice_url,
        protocol
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?)
RETURNING
    *;

-- name: DeleteKubernetes :exec
DELETE FROM
    kubernetes
WHERE
    name = ?;

-- name: GetKubernetesByName :one
SELECT
    *
FROM
    kubernetes
WHERE
    name = ?;

-- Docker queries
-- name: GetAllDocker :many
SELECT
    *
FROM
    docker;

-- name: InsertDocker :one
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
    (?, ?, ?, ?, ?, ?, ?, ?)
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

-- name: UpdateDocker :one
UPDATE
    docker
SET
    directory = ?,
    api_url = ?,
    gui_url = ?,
    backoffice_url = ?,
    api_port = ?,
    gui_port = ?,
    backoffice_port = ?
WHERE
    name = ?
RETURNING
    *;
