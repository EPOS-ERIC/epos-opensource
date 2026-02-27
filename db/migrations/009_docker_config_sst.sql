-- +goose Up
DROP TABLE docker;

CREATE TABLE docker (
    name TEXT NOT NULL PRIMARY KEY,
    config_yaml TEXT NOT NULL
);

-- +goose Down
DROP TABLE docker;

CREATE TABLE docker (
    name TEXT NOT NULL PRIMARY KEY,
    directory TEXT NOT NULL UNIQUE,
    api_url TEXT NOT NULL,
    gui_url TEXT NOT NULL,
    backoffice_url TEXT,
    api_port INTEGER NOT NULL,
    gui_port INTEGER NOT NULL,
    backoffice_port INTEGER
);
