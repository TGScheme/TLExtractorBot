-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS patched_objects (
    os              patch_os_enum NOT NULL,
    object_name     TEXT NOT NULL,
    old_constructor BIGINT NOT NULL,
    new_constructor BIGINT NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (os, object_name)
);
-- +goose StatementEnd
