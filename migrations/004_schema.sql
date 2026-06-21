-- +goose Up
CREATE TABLE IF NOT EXISTS table_meta (
    id          BIGSERIAL   PRIMARY KEY,
    space_id    BIGINT      NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    slug        TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(space_id, slug)
);

CREATE TABLE IF NOT EXISTS field_meta (
    id                   BIGSERIAL   PRIMARY KEY,
    table_id             BIGINT      NOT NULL REFERENCES table_meta(id) ON DELETE CASCADE,
    name                 TEXT        NOT NULL,
    slug                 TEXT        NOT NULL,
    type                 TEXT        NOT NULL,
    required             BOOLEAN     NOT NULL DEFAULT FALSE,
    default_value        TEXT,
    unique               BOOLEAN     NOT NULL DEFAULT FALSE,
    enum_values          TEXT[],
    relation_table_id    BIGINT      REFERENCES table_meta(id),
    relation_cardinality TEXT,
    position             INT         NOT NULL DEFAULT 0,
    UNIQUE(table_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_table_meta_space_id    ON table_meta(space_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_table_meta_slug ON table_meta(space_id, slug);
CREATE INDEX IF NOT EXISTS idx_field_meta_table_id    ON field_meta(table_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_field_meta_slug ON field_meta(table_id, slug);

-- +goose Down
DROP INDEX IF EXISTS idx_field_meta_slug;
DROP INDEX IF EXISTS idx_field_meta_table_id;
DROP INDEX IF EXISTS idx_table_meta_slug;
DROP INDEX IF EXISTS idx_table_meta_space_id;
DROP TABLE IF EXISTS field_meta;
DROP TABLE IF EXISTS table_meta;
