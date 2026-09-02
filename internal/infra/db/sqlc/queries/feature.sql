-- name: UpsertFeature :exec
INSERT INTO "Feature" (id, name)
VALUES (gen_random_uuid(), $1)
ON CONFLICT (name) DO NOTHING;

-- name: ListAllFeatures :many
SELECT id, name
FROM "Feature"
ORDER BY name;

-- name: GetFeatureByName :one
SELECT id, name
FROM "Feature"
WHERE name = $1
LIMIT 1;

-- name: GetFeatureByID :one
SELECT id, name
FROM "Feature"
WHERE id = $1
LIMIT 1;

-- name: CreateFeature :one
INSERT INTO "Feature" (id, name)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteFeature :exec
DELETE FROM "Feature"
WHERE id = $1;
