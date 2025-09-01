-- name: CreateRole :one
INSERT INTO "Role" (id, title, "enterpriseId")
VALUES ($1, $2, $3)
RETURNING *;
-- name: DeleteRole :exec
DELETE FROM "Role"
WHERE id = $1;
-- name: UpdateRole :one
UPDATE "Role"
SET title = $2
WHERE id = $1
RETURNING *;
-- name: ListRolesByEnterprise :many
SELECT "id",
  "title",
  "enterpriseId" as enterprise_id
FROM "Role"
WHERE "enterpriseId" = $1;
-- name: GetRoleByID :one
SELECT r."id" as role_id,
  r."title" as role_title,
  r."enterpriseId" as enterprise_id,
  COALESCE(
    json_agg(
      json_build_object(
        'id',
        p."id",
        'feature_id',
        p."featureId",
        'feature_name',
        f."name",
        'create',
        p."create",
        'read',
        p."read",
        'update',
        p."update",
        'delete',
        p."delete"
      )
      ORDER BY f."name"
    ) FILTER (
      WHERE p."id" IS NOT NULL
    ),
    '[]'::json
  ) as permissions
FROM "Role" r
  LEFT JOIN "Permission" p ON r."id" = p."roleId"
  LEFT JOIN "Feature" f ON p."featureId" = f."id"
WHERE r."id" = $1
GROUP BY r."id",
  r."title",
  r."enterpriseId";
