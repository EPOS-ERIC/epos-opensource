-- +goose Up
CREATE TABLE kubernetes (
    name TEXT NOT NULL PRIMARY KEY, -- just the name as the primary key to simplify cli usage (avoids having to pass name + context on each command)
    directory TEXT NOT NULL UNIQUE,
    context TEXT NOT NULL,
    api_url TEXT NOT NULL,
    gui_url TEXT NOT NULL,
    backoffice_url TEXT NOT NULL,
    protocol TEXT NOT NULL CHECK (protocol IN ('http', 'https'))
);

CREATE TABLE docker (
    name TEXT NOT NULL PRIMARY KEY,
    directory TEXT NOT NULL UNIQUE,
    api_url TEXT NOT NULL,
    gui_url TEXT NOT NULL,
    backoffice_url TEXT NOT NULL,
    api_port INTEGER NOT NULL,
    gui_port INTEGER NOT NULL,
    backoffice_port INTEGER NOT NULL
);

-- +goose Down
DROP TABLE docker;

DROP TABLE kubernetes;
