-- +goose Up
-- +goose StatementBegin
ALTER TABLE settings ADD COLUMN IF NOT EXISTS openai_model TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd
