-- +goose Up
CREATE TABLE IF NOT EXISTS pdf_templates (
    id              BIGSERIAL   PRIMARY KEY,
    space_id        BIGINT      NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT        NOT NULL,
    storage_path    TEXT        NOT NULL,
    placeholders    JSONB       NOT NULL DEFAULT '[]',
    status          TEXT        NOT NULL DEFAULT 'needs_mapping',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS pdf_mappings (
    id              BIGSERIAL   PRIMARY KEY,
    template_id     BIGINT      NOT NULL REFERENCES pdf_templates(id) ON DELETE CASCADE,
    placeholder     TEXT        NOT NULL,
    source_field_id BIGINT      REFERENCES field_meta(id),
    expression      TEXT,
    UNIQUE(template_id, placeholder)
);

CREATE TABLE IF NOT EXISTS pdf_jobs (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    space_id        BIGINT      NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    template_id     BIGINT      NOT NULL REFERENCES pdf_templates(id),
    status          TEXT        NOT NULL DEFAULT 'pending',
    total_records   INT,
    processed       INT         NOT NULL DEFAULT 0,
    storage_path    TEXT,
    created_by      BIGINT      NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pdf_templates_space_id ON pdf_templates(space_id);
CREATE INDEX IF NOT EXISTS idx_pdf_jobs_space_id      ON pdf_jobs(space_id);

-- +goose Down
DROP INDEX IF EXISTS idx_pdf_jobs_space_id;
DROP INDEX IF EXISTS idx_pdf_templates_space_id;
DROP TABLE IF EXISTS pdf_jobs;
DROP TABLE IF EXISTS pdf_mappings;
DROP TABLE IF EXISTS pdf_templates;
