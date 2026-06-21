-- +goose Up
CREATE TABLE IF NOT EXISTS spaces (
    id          BIGSERIAL   PRIMARY KEY,
    name        TEXT        NOT NULL,
    slug        TEXT        NOT NULL UNIQUE,
    owner_id    BIGINT      NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS space_members (
    space_id    BIGINT      NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        TEXT        NOT NULL DEFAULT 'member' CHECK (role IN ('admin','member')),
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (space_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_spaces_owner_id            ON spaces(owner_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_spaces_slug         ON spaces(slug);
CREATE UNIQUE INDEX IF NOT EXISTS idx_space_members_composite ON space_members(space_id, user_id);
CREATE INDEX IF NOT EXISTS idx_space_members_user_id      ON space_members(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_space_members_user_id;
DROP INDEX IF EXISTS idx_space_members_composite;
DROP INDEX IF EXISTS idx_spaces_slug;
DROP INDEX IF EXISTS idx_spaces_owner_id;
DROP TABLE IF EXISTS space_members;
DROP TABLE IF EXISTS spaces;
