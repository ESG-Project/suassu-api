package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/stretchr/testify/require"
)

type fakePermissionChecker struct {
	perms *types.UserPermissions
	err   error
}

func (f *fakePermissionChecker) GetUserPermissionsWithRole(ctx context.Context, userID, enterpriseID string) (*types.UserPermissions, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.perms, nil
}

func withClaims(claims appauth.Claims) *http.Request {
	req := httptest.NewRequest("GET", "/", nil)
	ctx := middleware.WithClaims(req.Context(), claims)
	return req.WithContext(ctx)
}

func okHandler(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

func TestRequirePermission(t *testing.T) {
	t.Parallel()

	checker := &fakePermissionChecker{perms: &types.UserPermissions{
		Permissions: []*types.UserPermission{{FeatureName: "Role", Create: true}},
	}}

	t.Run("allows when the feature/action flag is set", func(t *testing.T) {
		h := middleware.RequirePermission(checker, "Role", "create")(http.HandlerFunc(okHandler))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, withClaims(appauth.Claims{Subject: "u1", EnterpriseID: "e1"}))
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("denies when the flag is not set", func(t *testing.T) {
		h := middleware.RequirePermission(checker, "Role", "delete")(http.HandlerFunc(okHandler))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, withClaims(appauth.Claims{Subject: "u1", EnterpriseID: "e1"}))
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("requires authentication", func(t *testing.T) {
		h := middleware.RequirePermission(checker, "Role", "create")(http.HandlerFunc(okHandler))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequirePermissionAny(t *testing.T) {
	t.Parallel()

	t.Run("allows when actor has the action on any of the listed features", func(t *testing.T) {
		checker := &fakePermissionChecker{perms: &types.UserPermissions{
			Permissions: []*types.UserPermission{{FeatureName: "Financial", Create: true}},
		}}
		h := middleware.RequirePermissionAny(checker, "create", "Client", "Technician", "User", "Financial")(http.HandlerFunc(okHandler))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, withClaims(appauth.Claims{Subject: "u1", EnterpriseID: "e1"}))
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("denies when actor has none of the listed features", func(t *testing.T) {
		checker := &fakePermissionChecker{perms: &types.UserPermissions{
			Permissions: []*types.UserPermission{{FeatureName: "Project", Create: true}},
		}}
		h := middleware.RequirePermissionAny(checker, "create", "Client", "Technician", "User", "Financial")(http.HandlerFunc(okHandler))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, withClaims(appauth.Claims{Subject: "u1", EnterpriseID: "e1"}))
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("propagates checker errors", func(t *testing.T) {
		checker := &fakePermissionChecker{err: context.DeadlineExceeded}
		h := middleware.RequirePermissionAny(checker, "create", "Client")(http.HandlerFunc(okHandler))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, withClaims(appauth.Claims{Subject: "u1", EnterpriseID: "e1"}))
		require.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("requires authentication", func(t *testing.T) {
		checker := &fakePermissionChecker{}
		h := middleware.RequirePermissionAny(checker, "create", "Client")(http.HandlerFunc(okHandler))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
