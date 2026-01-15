package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	domain "github.com/ESG-Project/suassu-api/internal/domain/refreshtoken"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type RefreshTokenRepo struct {
	q *sqlc.Queries
}

// NewRefreshTokenRepo cria um novo repositório de refresh tokens.
func NewRefreshTokenRepo(db *sql.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{q: sqlc.New(db)}
}

// NewRefreshTokenRepoFrom cria um repositório a partir de um DBTX (para transações).
func NewRefreshTokenRepoFrom(d dbtx) *RefreshTokenRepo {
	return &RefreshTokenRepo{q: sqlc.New(d)}
}

// Create insere um novo refresh token no banco.
func (r *RefreshTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	return r.q.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
	})
}

// GetByHash busca um refresh token pelo hash.
func (r *RefreshTokenRepo) GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "refresh token not found")
		}
		return nil, err
	}

	token := &domain.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}

	if row.RevokedAt.Valid {
		revokedAt := row.RevokedAt.Time
		token.RevokedAt = &revokedAt
	}

	return token, nil
}

// Revoke revoga um refresh token específico.
func (r *RefreshTokenRepo) Revoke(ctx context.Context, id string) error {
	return r.q.RevokeRefreshToken(ctx, id)
}

// RevokeAllForUser revoga todos os refresh tokens de um usuário.
func (r *RefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	return r.q.RevokeAllUserRefreshTokens(ctx, userID)
}

// DeleteExpired remove tokens expirados e revogados.
func (r *RefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	return r.q.DeleteExpiredRefreshTokens(ctx)
}

// GetUserByRefreshToken busca o usuário associado ao refresh token.
func (r *RefreshTokenRepo) GetTokenIfValid(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	token, err := r.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if !token.IsValid() {
		if token.IsExpired() {
			return nil, apperr.New(apperr.CodeUnauthorized, "refresh token expired")
		}
		return nil, apperr.New(apperr.CodeUnauthorized, "refresh token revoked")
	}

	return token, nil
}

// RotateToken revoga o token antigo e cria um novo (atomic operation).
func (r *RefreshTokenRepo) RotateToken(ctx context.Context, oldTokenID string, newToken *domain.RefreshToken) error {
	// Revogar token antigo
	if err := r.Revoke(ctx, oldTokenID); err != nil {
		return err
	}

	// Criar novo token
	return r.Create(ctx, newToken)
}

// CleanupOldTokens remove tokens com mais de X dias.
func (r *RefreshTokenRepo) CleanupOldTokens(ctx context.Context, olderThan time.Duration) error {
	return r.DeleteExpired(ctx)
}
