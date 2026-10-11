-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dropped_objects (
    crc32      BIGINT PRIMARY KEY,
    build      BIGINT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd
-- +goose StatementBegin
DELETE FROM extracted_objects WHERE source = 'android';
-- +goose StatementEnd
