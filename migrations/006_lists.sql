-- +goose Up
CREATE TABLE IF NOT EXISTS lists (
    id              BIGSERIAL   PRIMARY KEY,
    space_id        BIGINT      NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    slug            TEXT        NOT NULL,
    source_table_id BIGINT      NOT NULL REFERENCES table_meta(id) ON DELETE CASCADE,
    field_config    JSONB       NOT NULL DEFAULT '[]',
    filter_config   JSONB       NOT NULL DEFAULT '{}',
    sort_config     JSONB       NOT NULL DEFAULT '[]',
    row_limit       INT         NOT NULL DEFAULT 100,
    published_at    TIMESTAMPTZ,
    unpublished_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(space_id, slug)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_lists_space_slug ON lists(space_id, slug);

-- +goose Down
DROP INDEX IF EXISTS idx_lists_space_slug;
DROP TABLE IF EXISTS lists;
