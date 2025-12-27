-- +goose Up
ALTER TABLE kubernetes RENAME TO k8s;

-- +goose Down
ALTER TABLE k8s RENAME TO kubernetes;