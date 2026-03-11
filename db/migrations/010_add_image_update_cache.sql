-- +goose Up
CREATE TABLE image_update_cache (
    image_ref TEXT NOT NULL PRIMARY KEY,
    remote_digest TEXT NOT NULL,
    remote_created_at DATETIME,
    fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE image_update_cache;
