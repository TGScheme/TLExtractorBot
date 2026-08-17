-- name: GetSettings :one
SELECT last_version_code,
       last_tdesk_id,
       last_tdlib_id,
       last_corefork_layer,
       status_message_id,
       tdesktop_branch,
       llm_model
FROM settings
WHERE id = TRUE;

-- name: SetLastVersionCode :exec
UPDATE settings SET last_version_code = @last_version_code, updated_at = NOW() WHERE id = TRUE;

-- name: SetLastTDeskID :exec
UPDATE settings SET last_tdesk_id = @last_tdesk_id, updated_at = NOW() WHERE id = TRUE;

-- name: SetLastTDLibID :exec
UPDATE settings SET last_tdlib_id = @last_tdlib_id, updated_at = NOW() WHERE id = TRUE;

-- name: SetLastCoreForkLayer :exec
UPDATE settings SET last_corefork_layer = @last_corefork_layer, updated_at = NOW() WHERE id = TRUE;

-- name: SetStatusMessageID :exec
UPDATE settings SET status_message_id = @status_message_id, updated_at = NOW() WHERE id = TRUE;

-- name: SetTDesktopBranch :exec
UPDATE settings SET tdesktop_branch = @tdesktop_branch, updated_at = NOW() WHERE id = TRUE;

-- name: SetLLMModel :exec
UPDATE settings SET llm_model = @llm_model, updated_at = NOW() WHERE id = TRUE;
