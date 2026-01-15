package refreshtoken

import "time"

// RefreshToken representa um token de refresh no domínio.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

// NewRefreshToken cria um novo RefreshToken.
func NewRefreshToken(id, userID, tokenHash string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}

// IsExpired verifica se o token expirou.
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// IsRevoked verifica se o token foi revogado.
func (r *RefreshToken) IsRevoked() bool {
	return r.RevokedAt != nil
}

// IsValid verifica se o token é válido (não expirado e não revogado).
func (r *RefreshToken) IsValid() bool {
	return !r.IsExpired() && !r.IsRevoked()
}
