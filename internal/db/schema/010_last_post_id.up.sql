-- +goose Up
-- +goose StatementBegin
ALTER TABLE settings ADD COLUMN IF NOT EXISTS last_post_id BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd
