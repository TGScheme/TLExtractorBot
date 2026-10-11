-- name: ListDroppedObjects :many
SELECT crc32, build FROM dropped_objects;

-- name: SetDroppedObject :exec
INSERT INTO dropped_objects (crc32, build) VALUES (@crc32, @build)
ON CONFLICT (crc32) DO UPDATE SET build = EXCLUDED.build, updated_at = NOW();

-- name: DeleteDroppedObject :exec
DELETE FROM dropped_objects WHERE crc32 = @crc32;
