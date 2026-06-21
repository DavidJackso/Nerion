-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id    BIGSERIAL PRIMARY KEY,
    name  TEXT      NOT NULL,
    email TEXT      NOT NULL UNIQUE,
    role  TEXT      NOT NULL DEFAULT 'user',
    password_hash TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS users;
