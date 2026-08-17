-- +goose Up
-- +goose StatementBegin
ALTER TABLE settings ALTER COLUMN llm_model SET DEFAULT 'gemini-2.5-flash';
UPDATE settings SET llm_model = 'gemini-2.5-flash' WHERE llm_model = '';
-- +goose StatementEnd
