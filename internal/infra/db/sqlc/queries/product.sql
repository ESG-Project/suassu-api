-- name: CreateProduct :exec
INSERT INTO "Product" (
    "id",
    "name",
    "suggestedValue",
    "enterpriseId",
    "parameterId",
    "deliverable",
    "typeProductId",
    "isDefault"
  )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpdateProduct :exec
UPDATE "Product"
SET "name" = $2,
  "suggestedValue" = $3,
  "parameterId" = $4,
  "deliverable" = $5,
  "typeProductId" = $6,
  "isDefault" = $7
WHERE id = $1
  AND "enterpriseId" = $8;

-- name: GetProductByID :one
SELECT id,
  name,
  "suggestedValue",
  "enterpriseId",
  "parameterId",
  deliverable,
  "typeProductId",
  "isDefault"
FROM "Product"
WHERE id = $1
  AND "enterpriseId" = $2
LIMIT 1;

-- name: ListProductsByEnterprise :many
SELECT id,
  name,
  "suggestedValue",
  "enterpriseId",
  "parameterId",
  deliverable,
  "typeProductId",
  "isDefault"
FROM "Product"
WHERE "enterpriseId" = $1
ORDER BY name;

-- name: DeleteProduct :exec
DELETE FROM "Product"
WHERE id = $1
  AND "enterpriseId" = $2;
