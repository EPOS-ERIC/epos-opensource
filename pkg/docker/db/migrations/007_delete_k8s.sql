-- +goose Up
DROP TABLE k8s;

-- +goose Down
CREATE TABLE k8s (
    name TEXT NOT NULL PRIMARY KEY,
    directory TEXT NOT NULL UNIQUE,
    context TEXT NOT NULL,
    api_url TEXT NOT NULL,
    gui_url TEXT NOT NULL,
    backoffice_url TEXT,
    protocol TEXT NOT NULL CHECK (protocol IN ('http', 'https')),
    tls_enabled BOOLEAN NOT NULL DEFAULT FALSE
);
