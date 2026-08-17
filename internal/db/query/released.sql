-- name: ListReleasedLayers :many
SELECT DISTINCT layer FROM released_objects ORDER BY layer;

-- name: GetMaxReleasedLayer :one
SELECT COALESCE(MAX(layer), 0)::int AS layer FROM released_objects;

-- name: HasReleasedLayer :one
SELECT EXISTS (SELECT 1 FROM released_objects WHERE layer = @layer) AS present;

-- name: DeleteReleasedLayer :exec
DELETE FROM released_objects WHERE layer = @layer;

-- name: ListRemovedConstructors :many
WITH ordered AS (
    SELECT layer, LEAD(layer) OVER (ORDER BY layer) AS next_layer
    FROM (SELECT DISTINCT layer FROM released_objects) l
),
disappeared AS (
    SELECT DISTINCT r.crc32
    FROM ordered o
    JOIN released_objects r ON r.layer = o.layer
    WHERE o.next_layer IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM released_objects n
          WHERE n.layer = o.next_layer AND n.crc32 = r.crc32
      )
)
SELECT d.crc32
FROM disappeared d
WHERE NOT EXISTS (
    SELECT 1 FROM released_objects r
    WHERE r.crc32 = d.crc32
      AND r.layer = (SELECT MAX(layer) FROM released_objects)
);
