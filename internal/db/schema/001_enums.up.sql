-- +goose Up
-- +goose StatementBegin
CREATE TYPE tl_kind_enum AS ENUM ('constructor', 'method');

CREATE TYPE api_kind_enum AS ENUM ('main', 'e2e');

CREATE TYPE scheme_role_enum AS ENUM ('stable', 'preview');

CREATE TYPE patch_os_enum AS ENUM ('android', 'ios', 'tdesktop', 'tdlib');
-- +goose StatementEnd
