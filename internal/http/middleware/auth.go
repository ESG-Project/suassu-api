package middleware

import (
	"context"
	"net/http"
	"strings"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
)

type claimsKeyType string

const claimsKey claimsKeyType = "auth_claims"

func AuthJWT(tokens appauth.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if !strings.HasPrefix(authz, "Bearer ") {
				httperr.Handle(w, r, apperr.New(apperr.CodeUnauthorized, "bearer token required"))
				return
			}
			raw := strings.TrimPrefix(authz, "Bearer ")
			parsed, err := tokens.Parse(raw)
			if err != nil {
				httperr.Handle(w, r, apperr.New(apperr.CodeUnauthorized, "invalid token"))
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, parsed)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromCtx(ctx context.Context) (appauth.Claims, bool) {
	val := ctx.Value(claimsKey)
	if val == nil {
		return appauth.Claims{}, false
	}

	// Agora JWT.Parse retorna appauth.Claims diretamente
	if claims, ok := val.(appauth.Claims); ok {
		return claims, true
	}

	return appauth.Claims{}, false
}
