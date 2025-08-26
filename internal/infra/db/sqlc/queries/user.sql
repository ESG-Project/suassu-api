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
-- name: GetUserByEmail :one
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
WHERE email = $1;
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
ORDER BY name ASC
LIMIT $1 OFFSET $2;
