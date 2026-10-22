-- +goose NO TRANSACTION
-- +goose Up
ALTER TYPE scheme_role_enum ADD VALUE IF NOT EXISTS 'synced';
ALTER TYPE scheme_role_enum ADD VALUE IF NOT EXISTS 'synced_previous';
