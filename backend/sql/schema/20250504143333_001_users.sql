-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    updated_at TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
