-- +goose Up
-- +goose StatementBegin
ALTER TABLE layer_pages ADD COLUMN IF NOT EXISTS story TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd
