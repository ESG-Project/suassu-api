-- name: CreateProject :one
INSERT INTO "Project" (
    id,
    title,
    cnpj,
    activity,
    codram,
    "usefulArea",
    "totalArea",
    "pollutingPower",
    stage,
    situation,
    "addressId",
    "clientId",
    size
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetProjectByID :one
SELECT 
    p.id,
    p.title,
    p.cnpj,
    p.activity,
    p.codram,
    p."usefulArea",
    p."totalArea",
    p."pollutingPower",
    p.stage,
    p.situation,
    p."addressId",
    p."clientId",
    p.size,
    a."zipCode",
    a.state,
    a.city,
    a.neighborhood,
    a.street,
    a.num,
    a.latitude,
    a.longitude,
    a."addInfo"
FROM "Project" p
LEFT JOIN "Address" a ON p."addressId" = a.id
WHERE p.id = $1
LIMIT 1;

-- name: ListProjectsByClient :many
SELECT 
    p.id,
    p.title,
    p.cnpj,
    p.activity,
    p.stage,
    p.situation,
    p."clientId"
FROM "Project" p
WHERE p."clientId" = $1
ORDER BY p.title ASC;

-- name: ListAllProjects :many
SELECT 
    p.id,
    p.title,
    p.cnpj,
    p.activity,
    p.stage,
    p.situation,
    p."clientId"
FROM "Project" p
ORDER BY p.title ASC
LIMIT $1 OFFSET $2;

-- name: UpdateProject :exec
UPDATE "Project"
SET
    title = $2,
    cnpj = $3,
    activity = $4,
    codram = $5,
    "usefulArea" = $6,
    "totalArea" = $7,
    "pollutingPower" = $8,
    stage = $9,
    situation = $10,
    size = $11
WHERE id = $1;

-- name: DeleteProject :exec
DELETE FROM "Project"
WHERE id = $1;

