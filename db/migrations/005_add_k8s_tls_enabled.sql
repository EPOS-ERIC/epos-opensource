-- +goose Up
ALTER TABLE
    k8s
ADD
    COLUMN tls_enabled BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE
    k8s DROP COLUMN tls_enabled;
