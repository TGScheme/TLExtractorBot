-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS layer_pages (
    layer      INTEGER PRIMARY KEY,
    path       TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd
