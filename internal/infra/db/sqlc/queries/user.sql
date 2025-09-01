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
-- name: GetUserPermissionsWithRole :one
SELECT u.id as user_id,
  u.name as user_name,
  r.id as role_id,
  r.title as role_title,
  COALESCE(
    json_agg(
      json_build_object(
        'id',
        p.id,
        'feature_id',
        p."featureId",
        'feature_name',
        f.name,
        'create',
        p."create",
        'read',
        p."read",
        'update',
        p."update",
        'delete',
        p."delete"
      )
      ORDER BY f.name
    ) FILTER (
      WHERE p.id IS NOT NULL
    ),
    '[]'::json
  ) as permissions
FROM "User" u
  JOIN "Role" r ON u."roleId" = r.id
  LEFT JOIN "Permission" p ON r.id = p."roleId"
  LEFT JOIN "Feature" f ON p."featureId" = f.id
WHERE u.id = $1
  AND u."enterpriseId" = $2
GROUP BY u.id,
  u.name,
  r.id,
  r.title;
-- name: GetUserWithDetails :one
SELECT u.id as user_id,
  u.name as user_name,
  u.email as user_email,
  u.document as user_document,
  u.phone as user_phone,
  u."addressId" as user_address_id,
  u."roleId" as user_role_id,
  u."enterpriseId" as user_enterprise_id,
  r.title as role_title,
  e.id as enterprise_id,
  e.name as enterprise_name,
  e.cnpj as enterprise_cnpj,
  e.email as enterprise_email,
  e."fantasyName" as enterprise_fantasy_name,
  e.phone as enterprise_phone,
  e."addressId" as enterprise_address_id,
  a.id as address_id,
  a."zipCode" as address_zip_code,
  a.state as address_state,
  a.city as address_city,
  a.neighborhood as address_neighborhood,
  a.street as address_street,
  a.num as address_num,
  a.latitude as address_latitude,
  a.longitude as address_longitude,
  a."addInfo" as address_add_info,
  COALESCE(
    json_agg(
      json_build_object(
        'id',
        p.id,
        'feature_id',
        p."featureId",
        'feature_name',
        f.name,
        'create',
        p."create",
        'read',
        p."read",
        'update',
        p."update",
        'delete',
        p."delete"
      )
      ORDER BY f.name
    ) FILTER (
      WHERE p.id IS NOT NULL
    ),
    '[]'::json
  ) as permissions
FROM "User" u
  JOIN "Role" r ON u."roleId" = r.id
  JOIN "Enterprise" e ON u."enterpriseId" = e.id
  LEFT JOIN "Address" a ON u."addressId" = a.id
  LEFT JOIN "Permission" p ON r.id = p."roleId"
  LEFT JOIN "Feature" f ON p."featureId" = f.id
WHERE u.id = $1
  AND u."enterpriseId" = $2
GROUP BY u.id,
  u.name,
  u.email,
  u.document,
  u.phone,
  u."addressId",
  u."roleId",
  u."enterpriseId",
  r.title,
  e.id,
  e.name,
  e.cnpj,
  e.email,
  e."fantasyName",
  e.phone,
  e."addressId",
  a.id,
  a."zipCode",
  a.state,
  a.city,
  a.neighborhood,
  a.street,
  a.num,
  a.latitude,
  a.longitude,
  a."addInfo";
