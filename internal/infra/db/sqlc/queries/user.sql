-- name: CreateUser :exec
INSERT INTO "User" (
    "id",
    "name",
    "email",
    "password",
    "document",
    "phone",
    "addressId",
    "roleId",
    "enterpriseId"
  )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
-- name: ListUsers :many
SELECT u.id,
  u.name,
  u.email,
  u.password AS password_hash,
  u.document,
  u.phone,
  u."addressId" AS address_id,
  u."roleId" AS role_id,
  u."enterpriseId" AS enterprise_id,
  a."zipCode" AS zip_code,
  a.state,
  a.city,
  a.neighborhood,
  a.street,
  a.num,
  a.latitude,
  a.longitude,
  a."addInfo" AS add_info
FROM "User" u
JOIN "Address" a ON u."addressId" = a.id
WHERE "enterpriseId" = $1
  AND (
    u.email > $3
    OR (
      u.email = $3
      AND u.id > $4
    )
  )
ORDER BY u.email ASC,
  u.id ASC
LIMIT $2;
-- name: GetUserByEmailInTenant :one
SELECT id,
  name,
  email,
  password AS password_hash,
  document,
  phone,
  "addressId" AS address_id,
  "roleId" AS role_id,
  "enterpriseId" AS enterprise_id
FROM "User"
WHERE "enterpriseId" = $1
  AND email = $2
LIMIT 1;
-- name: GetUserByEmailForAuth :one
SELECT id,
  name,
  email,
  password AS password_hash,
  document,
  phone,
  "addressId" AS address_id,
  "roleId" AS role_id,
  "enterpriseId" AS enterprise_id
FROM "User"
WHERE email = $1
LIMIT 1;
