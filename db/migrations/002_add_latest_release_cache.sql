-- +goose Up
CREATE TABLE latest_release_cache (
    id INTEGER PRIMARY KEY DEFAULT 1,
    tag_name TEXT NOT NULL,
    fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE latest_release_cache;
