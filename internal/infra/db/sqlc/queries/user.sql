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
  AND (
    email > $3
    OR (
      email = $3
      AND id > $4
    )
  )
ORDER BY email ASC,
  id ASC
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
