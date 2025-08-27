package middleware

import (
	"context"
	"net/http"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
)

// chave de contexto tipada para evitar colisões
type enterpriseKeyType string

const enterpriseKey enterpriseKeyType = "enterprise_id"

// RequireEnterprise garante que há um enterpriseID válido nos claims
// (pressupõe que AuthJWT já rodou e colocou appauth.Claims no contexto).
func RequireEnterprise(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromCtx(r.Context())
		if !ok || claims.EnterpriseID == "" {
			httperr.Handle(w, r, apperr.New(apperr.CodeForbidden, "enterprise access required"))
			return
		}
		ctx := context.WithValue(r.Context(), enterpriseKey, claims.EnterpriseID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// EnterpriseID recupera o enterpriseID previamente setado pelo RequireEnterprise.
// Retorna "" se não estiver presente.
func EnterpriseID(ctx context.Context) string {
	if v, ok := ctx.Value(enterpriseKey).(string); ok {
		return v
	}
	return ""
}

// (Opcional) Helper para montar Claims no contexto em testes (sem passar pelo JWT real).
func WithClaims(ctx context.Context, c appauth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}
