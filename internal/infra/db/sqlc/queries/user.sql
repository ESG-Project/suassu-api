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
  LEFT JOIN "Address" a ON u."addressId" = a.id
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
-- name: FindUserByEmailForAuth :one
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
-- name: GetUserByID :one
SELECT u.id,
  u.name,
  u.email,
  u.password AS password_hash,
  u.document,
  u.phone,
  u."roleId" AS role_id,
  u."enterpriseId" AS enterprise_id,
  u."addressId" AS address_id,
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
  LEFT JOIN "Address" a ON u."addressId" = a.id
WHERE u."enterpriseId" = $1
  AND u.id = $2
LIMIT 1;

-- name: GetUserByIDForRefresh :one
-- Busca usuário por ID sem filtro de tenant (para refresh token)
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
WHERE id = $1
LIMIT 1;
-- name: UpdateUserEditable :exec
UPDATE "User"
SET name = $3,
  email = $4,
  phone = $5,
  "addressId" = $6,
  password = $7
WHERE id = $1
  AND "enterpriseId" = $2;
-- name: UpdateUserAdmin :exec
-- Update administrativo: também altera document e roleId (self-service via
-- /auth/me não pode, mas um admin editando outro usuário pode).
UPDATE "User"
SET name = $3,
  email = $4,
  phone = $5,
  "addressId" = $6,
  document = $7,
  "roleId" = $8
WHERE id = $1
  AND "enterpriseId" = $2;
-- name: DeleteUser :exec
DELETE FROM "User"
WHERE id = $1
  AND "enterpriseId" = $2;
-- name: GetUserByDocumentInEnterprise :one
SELECT id, name, email, document
FROM "User"
WHERE "enterpriseId" = $1
  AND document = $2
LIMIT 1;
-- name: GetPrimaryAdminUserID :one
-- O usuário administrador "primário" de uma empresa: aquele criado no
-- onboarding, cujo e-mail é o mesmo da empresa e cujo papel é Administrador.
-- Não pode ser editado/deletado por outro admin (ver DeleteUserService no
-- user-crud).
SELECT u.id
FROM "User" u
  JOIN "Role" r ON r.id = u."roleId"
  JOIN "Enterprise" e ON e.id = u."enterpriseId"
WHERE u."enterpriseId" = $1
  AND u.email = e.email
  AND r.title = 'Administrador'
LIMIT 1;
-- name: ListNonClientUsersByEnterprise :many
-- Lista usuários da empresa que não têm registro de Client associado
-- (equivalente a manyUsers no user-crud), opcionalmente excluindo um id
-- (o admin primário). Passe excludeId = '' para não excluir ninguém.
SELECT u.id,
  u.name,
  u.email,
  u.phone,
  u.document,
  u."enterpriseId" as enterprise_id,
  a."zipCode" as zip_code,
  a.state,
  a.city,
  a.neighborhood,
  a.street,
  a.num,
  r.id as role_id,
  r.title as role_title,
  t.id as technician_id,
  t."proRegister" as pro_register,
  t.graduation as technician_graduation,
  t.ctf as technician_ctf
FROM "User" u
  LEFT JOIN "Address" a ON a.id = u."addressId"
  LEFT JOIN "Role" r ON r.id = u."roleId"
  LEFT JOIN "Technician" t ON t."userId" = u.id
  LEFT JOIN "Client" c ON c."userId" = u.id
WHERE u."enterpriseId" = $1
  AND c.id IS NULL
  AND (sqlc.arg(exclude_id)::text = '' OR u.id != sqlc.arg(exclude_id)::text);
