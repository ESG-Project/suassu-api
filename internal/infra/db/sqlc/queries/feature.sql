-- name: UpsertFeature :exec
INSERT INTO "Feature" (id, name)
VALUES (gen_random_uuid(), $1)
ON CONFLICT (name) DO NOTHING;
