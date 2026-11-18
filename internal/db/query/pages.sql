-- name: GetLayerPagePath :one
SELECT path FROM layer_pages WHERE layer = @layer;

-- name: SetLayerPagePath :exec
INSERT INTO layer_pages (layer, path) VALUES (@layer, @path)
ON CONFLICT (layer) DO UPDATE SET path = EXCLUDED.path, updated_at = NOW();

-- name: GetLayerStory :one
SELECT story FROM layer_pages WHERE layer = @layer;

-- name: SetLayerStory :exec
UPDATE layer_pages SET story = @story, updated_at = NOW() WHERE layer = @layer;
