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
