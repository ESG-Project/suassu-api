-- name: CreateTypeProduct :exec
INSERT INTO "TypeProduct" (id, type, "enterpriseId")
VALUES ($1, $2, $3);

-- name: GetTypeProductByID :one
SELECT id,
  type,
  "enterpriseId"
FROM "TypeProduct"
WHERE id = $1
LIMIT 1;

-- name: ListTypeProductsByEnterprise :many
SELECT id,
  type,
  "enterpriseId"
FROM "TypeProduct"
WHERE "enterpriseId" = $1
ORDER BY type;

-- name: UpdateTypeProduct :exec
UPDATE "TypeProduct"
SET type = $2
WHERE id = $1
  AND "enterpriseId" = $3;

-- name: DeleteTypeProduct :exec
DELETE FROM "TypeProduct"
WHERE id = $1
  AND "enterpriseId" = $2;
