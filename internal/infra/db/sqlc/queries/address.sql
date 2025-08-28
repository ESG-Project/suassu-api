-- name: CreateAddress :exec
INSERT INTO "Address" (
    "id",
    "zipCode",
    "state",
    "city",
    "neighborhood",
    "street",
    "num",
    "latitude",
    "longitude",
    "addInfo"
  )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetAddressByID :one
SELECT *
FROM "Address"
WHERE id = $1
LIMIT 1;

-- name: UpdateAddress :exec
UPDATE "Address"
SET
  "zipCode" = $2,
  "state" = $3,
  "city" = $4,
  "neighborhood" = $5,
  "street" = $6,
  "num" = $7,
  "latitude" = $8,
  "longitude" = $9,
  "addInfo" = $10
WHERE id = $1;

-- name: ListAddresses :many
SELECT *
FROM "Address";
