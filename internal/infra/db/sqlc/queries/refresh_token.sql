-- name: CreateRefreshToken :exec
INSERT INTO refresh_token (id, user_id, token_hash, expires_at, created_at)
VALUES ($1, $2, $3, $4, now());

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
FROM refresh_token
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE refresh_token
SET revoked_at = now()
WHERE id = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_token
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_token
WHERE expires_at < now() OR revoked_at IS NOT NULL;
