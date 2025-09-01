-- name: CreatePermission :one
INSERT INTO "Permission" (
    id,
    "featureId",
    "roleId",
    "create",
    "read",
    "update",
    "delete"
  )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
-- name: DeletePermission :exec
DELETE FROM "Permission"
WHERE id = $1;
-- name: UpdatePermission :one
UPDATE "Permission"
SET "create" = $2,
  "read" = $3,
  "update" = $4,
  "delete" = $5
WHERE id = $1
RETURNING *;
-- name: ListPermissionsByRole :many
SELECT p."id",
  p."featureId" as feature_id,
  f."name" as feature_name,
  p."roleId" as role_id,
  p."create",
  p."read",
  p."update",
  p."delete"
FROM "Permission" p
  JOIN "Feature" f ON p."featureId" = f."id"
WHERE "roleId" = $1;
