-- name: GetChannelCursor :one
SELECT last_post_id FROM channel_cursors WHERE channel = @channel;

-- name: SetChannelCursor :exec
INSERT INTO channel_cursors (channel, last_post_id) VALUES (@channel, @last_post_id)
ON CONFLICT (channel) DO UPDATE SET last_post_id = EXCLUDED.last_post_id, updated_at = NOW();
