-- name: CreateLog :exec
INSERT INTO "Log" (id, tag, "enterpriseId", description, "createdAt")
VALUES ($1, $2, $3, $4, $5);
