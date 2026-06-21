-- +goose Up
CREATE TABLE IF NOT EXISTS api_keys (
    id            BIGSERIAL   PRIMARY KEY,
    space_id      BIGINT      NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name          TEXT        NOT NULL,
    key_hash      TEXT        NOT NULL UNIQUE,
    key_prefix    TEXT        NOT NULL,
    scope         TEXT        NOT NULL DEFAULT 'read_write' CHECK (scope IN ('read','read_write')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at  TIMESTAMPTZ,
    revoked_at    TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_hash    ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_space_id       ON api_keys(space_id);

-- +goose Down
DROP INDEX IF EXISTS idx_api_keys_space_id;
DROP INDEX IF EXISTS idx_api_keys_hash;
DROP TABLE IF EXISTS api_keys;
