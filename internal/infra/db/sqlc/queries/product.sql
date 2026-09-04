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

-- name: GetProductByIDAnyEnterprise :one
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
LIMIT 1;

-- name: ListProductsDetailedByEnterprise :many
SELECT p.id,
  p.name,
  p."suggestedValue",
  p."enterpriseId",
  p.deliverable,
  par.id AS parameter_id,
  par.title AS parameter_title,
  par.value AS parameter_value,
  tp.id AS type_product_id,
  tp.type AS type_product_type
FROM "Product" p
  LEFT JOIN "Parameter" par ON par.id = p."parameterId"
  LEFT JOIN "TypeProduct" tp ON tp.id = p."typeProductId"
WHERE p."enterpriseId" = $1
ORDER BY p.name;
