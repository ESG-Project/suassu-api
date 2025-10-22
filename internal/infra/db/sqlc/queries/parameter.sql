-- name: CreateParameter :exec
INSERT INTO "Parameter" (
    "id",
    "title",
    "value",
    "enterpriseId",
    "isDefault"
  )
VALUES ($1, $2, $3, $4, $5);

-- name: UpdateParameter :exec
UPDATE "Parameter"
SET "title" = $2,
  "value" = $3,
  "isDefault" = $4
WHERE id = $1
  AND "enterpriseId" = $5;

-- name: GetParameterByID :one
SELECT id,
  title,
  value,
  "enterpriseId",
  "isDefault"
FROM "Parameter"
WHERE id = $1
  AND "enterpriseId" = $2
LIMIT 1;

-- name: ListParametersByEnterprise :many
SELECT id,
  title,
  value,
  "enterpriseId",
  "isDefault"
FROM "Parameter"
WHERE "enterpriseId" = $1
ORDER BY title;

-- name: DeleteParameter :exec
DELETE FROM "Parameter"
WHERE id = $1
  AND "enterpriseId" = $2;
