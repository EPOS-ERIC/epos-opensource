-- +goose Up
CREATE TABLE docker_new (
    name TEXT NOT NULL PRIMARY KEY,
    directory TEXT NOT NULL UNIQUE,
    api_url TEXT NOT NULL,
    gui_url TEXT NOT NULL,
    backoffice_url TEXT,
    api_port INTEGER NOT NULL,
    gui_port INTEGER NOT NULL,
    backoffice_port INTEGER
);

INSERT INTO
    docker_new
SELECT
    *
FROM
    docker;

DROP TABLE docker;

ALTER TABLE
    docker_new RENAME TO docker;

CREATE TABLE k8s_new (
    name TEXT NOT NULL PRIMARY KEY,
    directory TEXT NOT NULL UNIQUE,
    context TEXT NOT NULL,
    api_url TEXT NOT NULL,
    gui_url TEXT NOT NULL,
    backoffice_url TEXT,
    protocol TEXT NOT NULL CHECK (protocol IN ('http', 'https')),
    tls_enabled BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO
    k8s_new
SELECT
    *
FROM
    k8s;

DROP TABLE k8s;

ALTER TABLE
    k8s_new RENAME TO k8s;

-- +goose Down
CREATE TABLE docker_new (
    name TEXT NOT NULL PRIMARY KEY,
    directory TEXT NOT NULL UNIQUE,
    api_url TEXT NOT NULL,
    gui_url TEXT NOT NULL,
    backoffice_url TEXT NOT NULL,
    api_port INTEGER NOT NULL,
    gui_port INTEGER NOT NULL,
    backoffice_port INTEGER NOT NULL
);

INSERT INTO
    docker_new
SELECT
    *
FROM
    docker;

DROP TABLE docker;

ALTER TABLE
    docker_new RENAME TO docker;

CREATE TABLE k8s_new (
    name TEXT NOT NULL PRIMARY KEY,
    directory TEXT NOT NULL UNIQUE,
    context TEXT NOT NULL,
    api_url TEXT NOT NULL,
    gui_url TEXT NOT NULL,
    backoffice_url TEXT NOT NULL,
    protocol TEXT NOT NULL CHECK (protocol IN ('http', 'https')),
    tls_enabled BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO
    k8s_new
SELECT
    *
FROM
    k8s;

DROP TABLE k8s;

ALTER TABLE
    k8s_new RENAME TO k8s;
