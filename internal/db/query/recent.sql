-- name: ListRecentLayers :many
SELECT layer FROM recent_layers ORDER BY created_at;

-- name: AddRecentLayer :exec
INSERT INTO recent_layers (layer) VALUES (@layer) ON CONFLICT DO NOTHING;

-- name: TrimRecentLayers :exec
DELETE FROM recent_layers
WHERE layer NOT IN (
    SELECT layer FROM recent_layers ORDER BY created_at DESC LIMIT @keep
);
