-- name: GetPatchedObject :one
SELECT old_constructor, new_constructor
FROM patched_objects
WHERE os = @os AND object_name = @object_name;

-- name: ListPatchedObjects :many
SELECT os, object_name, old_constructor, new_constructor FROM patched_objects;

-- name: UpsertPatchedObject :exec
INSERT INTO patched_objects (os, object_name, old_constructor, new_constructor)
VALUES (@os, @object_name, @old_constructor, @new_constructor)
ON CONFLICT (os, object_name) DO UPDATE
    SET old_constructor = EXCLUDED.old_constructor,
        new_constructor = EXCLUDED.new_constructor,
        updated_at = NOW();

-- name: DeletePatchedObject :exec
DELETE FROM patched_objects WHERE os = @os AND object_name = @object_name;

-- name: DeleteAllPatchedObjects :exec
DELETE FROM patched_objects;
