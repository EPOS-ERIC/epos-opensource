PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS environment (
    name TEXT NOT NULL,
    directory TEXT NOT NULL,
    platform TEXT CHECK (platform IN ('kubernetes', 'docker')) NOT NULL,
    PRIMARY KEY(name, platform)
);
