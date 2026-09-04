-- name: GetBankByCode :one
SELECT id,
  code,
  name
FROM "Bank"
WHERE code = $1
LIMIT 1;

-- name: CreateBank :exec
INSERT INTO "Bank" (id, code, name)
VALUES ($1, $2, $3);

-- name: GetEnterpriseBankByBankAndEnterprise :one
SELECT id,
  "enterpriseId",
  "bankId"
FROM "EnterpriseBank"
WHERE "bankId" = $1
  AND "enterpriseId" = $2
LIMIT 1;

-- name: GetEnterpriseBankByID :one
SELECT eb.id,
  eb."enterpriseId",
  eb."bankId",
  e.name AS enterprise_name
FROM "EnterpriseBank" eb
  JOIN "Enterprise" e ON e.id = eb."enterpriseId"
WHERE eb.id = $1
LIMIT 1;

-- name: CreateEnterpriseBank :exec
INSERT INTO "EnterpriseBank" (id, "enterpriseId", "bankId")
VALUES ($1, $2, $3);

-- name: DeleteEnterpriseBank :exec
DELETE FROM "EnterpriseBank"
WHERE id = $1;

-- name: ListEnterpriseBanksByEnterprise :many
SELECT eb.id,
  b.id AS bank_id,
  b.code AS bank_code,
  b.name AS bank_name
FROM "EnterpriseBank" eb
  JOIN "Bank" b ON b.id = eb."bankId"
WHERE eb."enterpriseId" = $1
ORDER BY b.name;
