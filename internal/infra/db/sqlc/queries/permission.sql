-- name: CreatePermission :one
INSERT INTO "Permission" (id, "featureId", "roleId", "create", "read", "update", "delete")
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM "Permission"
WHERE id = $1;

-- name: UpdatePermission :one
UPDATE "Permission"
SET "create" = $2, "read" = $3, "update" = $4, "delete" = $5
WHERE id = $1
RETURNING *;

-- name: ListPermissionsByRole :many
SELECT *
FROM "Permission"
WHERE "roleId" = $1;
