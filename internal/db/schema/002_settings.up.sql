-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS settings (
    id                  BOOLEAN PRIMARY KEY DEFAULT TRUE,
    last_version_code   BIGINT NOT NULL DEFAULT 0,
    last_tdesk_id       BIGINT NOT NULL DEFAULT 0,
    last_tdlib_id       BIGINT NOT NULL DEFAULT 0,
    last_corefork_layer INTEGER NOT NULL DEFAULT 0,
    status_message_id   BIGINT NOT NULL DEFAULT 0,
    tdesktop_branch     TEXT NOT NULL DEFAULT 'dev',
    llm_model           TEXT NOT NULL DEFAULT '',
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT settings_singleton CHECK (id)
);

INSERT INTO settings (id) VALUES (TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd
