-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS channel_cursors (
    channel      TEXT PRIMARY KEY,
    last_post_id BIGINT NOT NULL,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'settings' AND column_name = 'last_post_id'
    ) THEN
        INSERT INTO channel_cursors (channel, last_post_id)
        SELECT 'TAndroidBeta', last_post_id FROM settings WHERE id = TRUE AND last_post_id > 0
        ON CONFLICT (channel) DO NOTHING;
        ALTER TABLE settings DROP COLUMN last_post_id;
    END IF;
END $$;
-- +goose StatementEnd
