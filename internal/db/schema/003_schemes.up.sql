-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS schemes (
    id         BIGSERIAL PRIMARY KEY,
    layer      INTEGER NOT NULL,
    role       scheme_role_enum,
    is_sync    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_schemes_role
    ON schemes (role) WHERE role IS NOT NULL;

CREATE TABLE IF NOT EXISTS tl_objects (
    id           BIGSERIAL PRIMARY KEY,
    scheme_id    BIGINT NOT NULL REFERENCES schemes (id) ON DELETE CASCADE,
    api          api_kind_enum NOT NULL,
    kind         tl_kind_enum NOT NULL,
    crc32        BIGINT NOT NULL,
    object_name  TEXT NOT NULL,
    result       TEXT NOT NULL,
    layer        INTEGER NOT NULL DEFAULT 0,
    force_secret BOOLEAN NOT NULL DEFAULT FALSE,
    position     INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tl_objects_name
    ON tl_objects (scheme_id, api, kind, object_name);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tl_objects_position
    ON tl_objects (scheme_id, position);

CREATE INDEX IF NOT EXISTS idx_tl_objects_crc32
    ON tl_objects (scheme_id, crc32);

CREATE TABLE IF NOT EXISTS tl_params (
    object_id BIGINT NOT NULL REFERENCES tl_objects (id) ON DELETE CASCADE,
    position  INTEGER NOT NULL,
    name      TEXT NOT NULL,
    param_type TEXT NOT NULL,
    PRIMARY KEY (object_id, position)
);
-- +goose StatementEnd
