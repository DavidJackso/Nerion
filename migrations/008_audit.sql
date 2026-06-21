-- +goose Up
CREATE TABLE IF NOT EXISTS audit_log (
    id          BIGSERIAL   PRIMARY KEY,
    space_id    BIGINT      REFERENCES spaces(id) ON DELETE SET NULL,
    user_id     BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    action      TEXT        NOT NULL,
    entity_type TEXT,
    entity_id   TEXT,
    meta        JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_space_created ON audit_log(space_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_action        ON audit_log(space_id, action);

-- +goose Down
DROP INDEX IF EXISTS idx_audit_log_action;
DROP INDEX IF EXISTS idx_audit_log_space_created;
DROP TABLE IF EXISTS audit_log;
