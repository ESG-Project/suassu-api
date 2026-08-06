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
