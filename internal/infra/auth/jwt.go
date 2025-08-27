package auth

import (
	"time"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/config"
	domain "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

func NewJWT(cfg *config.Config) *JWT {
	return &JWT{
		secret:   []byte(cfg.JWTSecret),
		issuer:   cfg.JWTIssuer,
		audience: cfg.JWTAudience,
		ttl:      time.Duration(cfg.JWTAccessTTLMin) * time.Minute,
	}
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
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
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
