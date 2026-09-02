-- name: CreateTechnician :one
INSERT INTO "Technician" (id, "proRegister", graduation, ctf, "userId")
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTechnicianByUserID :one
SELECT id, "proRegister", graduation, ctf, "userId"
FROM "Technician"
WHERE "userId" = $1
LIMIT 1;

-- name: UpsertTechnicianByUserID :one
INSERT INTO "Technician" (id, "proRegister", graduation, ctf, "userId")
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT ("userId") DO UPDATE SET
  "proRegister" = EXCLUDED."proRegister",
  graduation = EXCLUDED.graduation,
  ctf = EXCLUDED.ctf
RETURNING *;
