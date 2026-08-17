-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS recent_layers (
    layer      INTEGER PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd
