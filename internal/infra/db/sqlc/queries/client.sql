-- name: CreateClient :one
INSERT INTO "Client" (id, "fantasyName", "userId")
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetClientByUserID :one
SELECT id, "fantasyName", "userId"
FROM "Client"
WHERE "userId" = $1
LIMIT 1;

-- name: UpdateClientByUserID :one
UPDATE "Client"
SET "fantasyName" = $2
WHERE "userId" = $1
RETURNING *;

-- name: UpsertClientByUserID :one
INSERT INTO "Client" (id, "fantasyName", "userId")
VALUES ($1, $2, $3)
ON CONFLICT ("userId") DO UPDATE SET "fantasyName" = EXCLUDED."fantasyName"
RETURNING *;

-- name: ListClientsByEnterprise :many
SELECT c.id AS client_id,
  c."fantasyName" AS fantasy_name,
  u.id,
  u.name,
  u.email,
  u.phone,
  u.document,
  u."enterpriseId" AS enterprise_id,
  a."zipCode" AS zip_code,
  a.state,
  a.city,
  a.neighborhood,
  a.street,
  a.num,
  r.id AS role_id,
  r.title AS role_title
FROM "Client" c
  JOIN "User" u ON u.id = c."userId"
  LEFT JOIN "Address" a ON a.id = u."addressId"
  LEFT JOIN "Role" r ON r.id = u."roleId"
WHERE u."enterpriseId" = $1;
