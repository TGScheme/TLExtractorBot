-- name: ListExtractedObjects :many
SELECT crc32 FROM extracted_objects WHERE source = @source;

-- name: DeleteExtractedObjects :exec
DELETE FROM extracted_objects WHERE source = @source;
