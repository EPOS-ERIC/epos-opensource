PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS kubernetes (
    name TEXT NOT NULL PRIMARY KEY,
    directory TEXT NOT NULL UNIQUE,
    context TEXT NOT NULL,
    api_url TEXT NOT NULL,
    gui_url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS docker (
    name TEXT NOT NULL PRIMARY KEY,
    directory TEXT NOT NULL UNIQUE,
    api_url TEXT NOT NULL,
    gui_url TEXT NOT NULL
);
