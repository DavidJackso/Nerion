-- +goose Up
CREATE TABLE IF NOT EXISTS space_invitations (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    space_id    BIGINT      NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    email       TEXT        NOT NULL,
    invited_by  BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_space_invitations_email ON space_invitations(email);

-- +goose Down
DROP INDEX IF EXISTS idx_space_invitations_email;
DROP TABLE IF EXISTS space_invitations;
