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
SET "featureId" = $2,
  "roleId" = $3,
  "create" = $4,
  "read" = $5,
  "update" = $6,
  "delete" = $7
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
-- name: GetPermissionByID :one
SELECT p."id",
  p."featureId" as feature_id,
  p."roleId" as role_id,
  p."create",
  p."read",
  p."update",
  p."delete",
  r."enterpriseId" as enterprise_id
FROM "Permission" p
  JOIN "Role" r ON p."roleId" = r."id"
WHERE p."id" = $1
LIMIT 1;
-- name: ListPermissionsByEnterprise :many
SELECT p."id",
  p."featureId" as feature_id,
  p."roleId" as role_id,
  p."create",
  p."read",
  p."update",
  p."delete"
FROM "Permission" p
  JOIN "Role" r ON p."roleId" = r."id"
WHERE r."enterpriseId" = $1;
