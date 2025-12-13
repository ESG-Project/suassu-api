-- name: CreateEnterprise :exec
INSERT INTO "Enterprise" (
    "id",
    "cnpj",
    "addressId",
    "email",
    "fantasyName",
    "name",
    "phone"
  )
VALUES ($1, $2, $3, $4, $5, $6, $7);
-- name: UpdateEnterprise :exec
UPDATE "Enterprise"
SET "cnpj" = $2,
  "addressId" = $3,
  "email" = $4,
  "fantasyName" = $5,
  "name" = $6,
  "phone" = $7
WHERE id = $1;
-- name: GetEnterpriseByID :one
SELECT e.id,
  e.cnpj,
  e.email,
  e.name,
  e."fantasyName",
  e.phone,
  e."addressId",
  a."zipCode" AS zip_code,
  a.state,
  a.city,
  a.neighborhood,
  a.street,
  a.num,
  a.latitude,
  a.longitude,
  a."addInfo" AS add_info,
  -- agrega products como JSON ([] quando não houver)
  COALESCE(
    json_agg(
      DISTINCT jsonb_build_object(
        'id', p.id,
        'name', p.name,
        'suggestedValue', p."suggestedValue",
        'enterpriseId', p."enterpriseId",
        'parameterId', p."parameterId",
        'deliverable', p.deliverable,
        'typeProductId', p."typeProductId",
        'isDefault', p."isDefault"
      )
    ) FILTER (WHERE p.id IS NOT NULL),
    '[]'
  ) AS products,
  -- agrega parameters como JSON ([] quando não houver)
  COALESCE(
    json_agg(
      DISTINCT jsonb_build_object(
        'id', pr.id,
        'title', pr.title,
        'value', pr.value,
        'enterpriseId', pr."enterpriseId",
        'isDefault', pr."isDefault"
      )
    ) FILTER (WHERE pr.id IS NOT NULL),
    '[]'
  ) AS parameters
FROM "Enterprise" e
  JOIN "Address" a ON e."addressId" = a.id
  LEFT JOIN "Product" p ON p."enterpriseId" = e.id
  LEFT JOIN "Parameter" pr ON pr."enterpriseId" = e.id
WHERE e.id = $1
GROUP BY e.id,
  a.id
LIMIT 1;
