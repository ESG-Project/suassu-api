-- name: CreateEnterprise :exec
INSERT INTO "Enterprise" (
    "id",
    "cnpj",
    "addressId",
    "email",
    "fantasyName",
    "name",
    "phone"
  )
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetEnterpriseByID :one
SELECT
  e.id,
  e.cnpj,
  e.email,
  e.name,
  e."fantasyName",
  e.phone,
  e."addressId",
  a."zipCode" AS zip_code,
  a.state,
  a.city,
  a.neighborhood,
  a.street,
  a.num,
  a.latitude,
  a.longitude,
  a."addInfo" AS add_info
FROM "Enterprise" e
JOIN "Address" a ON e."addressId" = a.id
WHERE e.id = $1
LIMIT 1;

-- name: UpdateEnterprise :exec
UPDATE "Enterprise"
SET
  "cnpj" = $2,
  "addressId" = $3,
  "email" = $4,
  "fantasyName" = $5,
  "name" = $6,
  "phone" = $7
WHERE id = $1;
