-- name: CreateRole :one
INSERT INTO "Role" (id, title, "enterpriseId")
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM "Role"
WHERE id = $1;

-- name: UpdateRole :one
UPDATE "Role"
SET title = $2
WHERE id = $1
RETURNING *;

-- name: ListRolesByEnterprise :many
SELECT *
FROM "Role"
WHERE "enterpriseId" = $1;
