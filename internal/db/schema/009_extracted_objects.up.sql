-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS extracted_objects (
    source TEXT NOT NULL,
    crc32  BIGINT NOT NULL,
    PRIMARY KEY (source, crc32)
);
-- +goose StatementEnd
