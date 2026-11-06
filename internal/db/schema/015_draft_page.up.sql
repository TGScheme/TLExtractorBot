-- +goose Up
-- +goose StatementBegin
ALTER TABLE settings ADD COLUMN IF NOT EXISTS draft_page_key TEXT NOT NULL DEFAULT '';
ALTER TABLE settings ADD COLUMN IF NOT EXISTS draft_page_path TEXT NOT NULL DEFAULT '';
UPDATE settings SET status_message_id = 0;
-- +goose StatementEnd
