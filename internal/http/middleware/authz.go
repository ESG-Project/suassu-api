package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
)

// PermissionChecker resolve as permissões (feature CRUD) do usuário logado.
// Implementado pelo serviço de usuário (GetUserPermissionsWithRole).
type PermissionChecker interface {
	GetUserPermissionsWithRole(ctx context.Context, userID, enterpriseID string) (*types.UserPermissions, error)
}

// RequirePermission garante que o papel do usuário possui a ação CRUD indicada
// (create/read/update/delete) sobre a feature informada. Pressupõe que AuthJWT e
// RequireEnterprise já rodaram.
func RequirePermission(checker PermissionChecker, feature, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromCtx(r.Context())
			if !ok || claims.Subject == "" {
				httperr.Handle(w, r, apperr.New(apperr.CodeUnauthorized, "authentication required"))
				return
			}

			perms, err := checker.GetUserPermissionsWithRole(r.Context(), claims.Subject, claims.EnterpriseID)
			if err != nil {
				httperr.Handle(w, r, err)
				return
			}

			if !hasPermission(perms, feature, action) {
				httperr.Handle(w, r, apperr.New(apperr.CodeForbidden, "insufficient permissions"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermissionAny garante que o papel do usuário possui a ação CRUD
// indicada em ao menos uma das features informadas ("qualquer uma destas").
// Equivalente ao suporte a array de features em permissionVerification no
// user-crud (ex.: POST /user aceita quem pode criar Client OU Technician OU
// User OU Financial).
func RequirePermissionAny(checker PermissionChecker, action string, features ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromCtx(r.Context())
			if !ok || claims.Subject == "" {
				httperr.Handle(w, r, apperr.New(apperr.CodeUnauthorized, "authentication required"))
				return
			}

			perms, err := checker.GetUserPermissionsWithRole(r.Context(), claims.Subject, claims.EnterpriseID)
			if err != nil {
				httperr.Handle(w, r, err)
				return
			}

			for _, feature := range features {
				if hasPermission(perms, feature, action) {
					next.ServeHTTP(w, r)
					return
				}
			}

			httperr.Handle(w, r, apperr.New(apperr.CodeForbidden, "insufficient permissions"))
		})
	}
}

func hasPermission(perms *types.UserPermissions, feature, action string) bool {
	if perms == nil {
		return false
	}
	for _, p := range perms.Permissions {
		if p == nil || !strings.EqualFold(p.FeatureName, feature) {
			continue
		}
		switch strings.ToLower(action) {
		case "create":
			return p.Create
		case "read":
			return p.Read
		case "update":
			return p.Update
		case "delete":
			return p.Delete
		}
	}
	return false
}

// Actor devolve o id do usuário autenticado e a empresa do contexto de uma
// vez — o par que os handlers de escrita precisam para aplicar o escopo por
// empresa e registrar a trilha de auditoria. ok é false quando não há claims
// no contexto (rota sem AuthJWT).
func Actor(ctx context.Context) (actorID, enterpriseID string, ok bool) {
	claims, ok := ClaimsFromCtx(ctx)
	if !ok {
		return "", "", false
	}
	return claims.Subject, EnterpriseID(ctx), true
}
