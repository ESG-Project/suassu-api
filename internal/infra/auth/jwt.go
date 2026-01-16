package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/config"
	domain "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret        []byte
	issuer        string
	audience      string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJWT(cfg *config.Config) *JWT {
	return &JWT{
		secret:        []byte(cfg.JWTSecret),
		issuer:        cfg.JWTIssuer,
		audience:      cfg.JWTAudience,
		accessTTL:     time.Duration(cfg.JWTAccessTTLMin) * time.Minute,
		refreshTTL:    time.Duration(cfg.JWTRefreshTTLDays) * 24 * time.Hour,
	}
}

// GetAccessTTL retorna o TTL do access token em segundos.
func (j *JWT) GetAccessTTL() int64 {
	return int64(j.accessTTL.Seconds())
}

// GetRefreshTTL retorna o TTL do refresh token.
func (j *JWT) GetRefreshTTL() time.Duration {
	return j.refreshTTL
}

type customClaims struct {
	Email        string  `json:"email"`
	Name         string  `json:"name"`
	EnterpriseID string  `json:"enterprise_id"`
	RoleID       *string `json:"role_id,omitempty"`
	jwt.RegisteredClaims
}

func (j *JWT) NewAccessToken(u *domain.User) (string, error) {
	now := time.Now().UTC()
	claims := customClaims{
		Email:        u.Email,
		Name:         u.Name,
		EnterpriseID: u.EnterpriseID,
		RoleID:       u.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			Issuer:    j.issuer,
			Audience:  []string{j.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// NewRefreshToken gera um novo refresh token (string aleatória).
func (j *JWT) NewRefreshToken() (token string, hash string, expiresAt time.Time, err error) {
	// Gerar 32 bytes aleatórios
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return "", "", time.Time{}, err
	}

	// Codificar em base64 URL-safe
	token = base64.URLEncoding.EncodeToString(bytes)

	// Criar hash SHA-256 para armazenamento
	hashBytes := sha256.Sum256([]byte(token))
	hash = hex.EncodeToString(hashBytes[:])

	// Calcular expiração
	expiresAt = time.Now().UTC().Add(j.refreshTTL)

	return token, hash, expiresAt, nil
}

// HashRefreshToken cria um hash SHA-256 de um refresh token.
func HashRefreshToken(token string) string {
	hashBytes := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hashBytes[:])
}

func (j *JWT) Parse(tokenStr string) (appauth.Claims, error) {
	var out appauth.Claims
	t, err := jwt.ParseWithClaims(tokenStr, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	}, jwt.WithAudience(j.audience), jwt.WithIssuer(j.issuer))
	if err != nil {
		return out, err
	}
	if !t.Valid {
		return out, jwt.ErrTokenInvalidClaims
	}
	cc := t.Claims.(*customClaims)
	out = appauth.Claims{
		Subject:      cc.Subject,
		Email:        cc.Email,
		EnterpriseID: cc.EnterpriseID,
		RoleID:       cc.RoleID,
		Name:         cc.Name,
	}
	return out, nil
}
