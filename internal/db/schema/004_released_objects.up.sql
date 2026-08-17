-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS released_objects (
    layer INTEGER NOT NULL,
    kind  tl_kind_enum NOT NULL,
    crc32 BIGINT NOT NULL,
    PRIMARY KEY (layer, kind, crc32)
);

CREATE INDEX IF NOT EXISTS idx_released_objects_crc32
    ON released_objects (crc32);
-- +goose StatementEnd
