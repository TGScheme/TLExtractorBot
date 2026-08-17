-- name: GetSchemeByRole :one
SELECT id, layer, is_sync
FROM schemes
WHERE role = @role;

-- name: CreateScheme :one
INSERT INTO schemes (layer, is_sync)
VALUES (@layer, @is_sync)
RETURNING id;

-- name: ClearSchemeRole :exec
UPDATE schemes SET role = NULL WHERE role = @role;

-- name: SetSchemeRole :exec
UPDATE schemes SET role = @role WHERE id = @id;

-- name: DeleteUnreferencedSchemes :exec
DELETE FROM schemes WHERE role IS NULL;

-- name: GetSchemeObjects :many
SELECT o.id, o.api, o.kind, o.crc32, o.object_name, o.result, o.layer, o.force_secret
FROM tl_objects o
WHERE o.scheme_id = @scheme_id
ORDER BY o.api, o.kind, o.position;

-- name: GetSchemeParams :many
SELECT p.object_id, p.name, p.param_type
FROM tl_params p
JOIN tl_objects o ON o.id = p.object_id
WHERE o.scheme_id = @scheme_id
ORDER BY p.object_id, p.position;
