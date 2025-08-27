package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/stretchr/testify/require"
)

func TestRequireEnterprise(t *testing.T) {
	t.Run("with valid enterprise ID", func(t *testing.T) {
		claims := appauth.Claims{
			Subject:      "user-123",
			Email:        "test@example.com",
			EnterpriseID: "ent-1",
			Name:         "Test User",
		}
		ctx := WithClaims(context.Background(), claims)

		req := httptest.NewRequest("GET", "/test", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		called := false
		handler := RequireEnterprise(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			enterpriseID := EnterpriseID(r.Context())
			require.Equal(t, "ent-1", enterpriseID)
		}))

		handler.ServeHTTP(w, req)

		require.True(t, called)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("without claims", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		called := false
		handler := RequireEnterprise(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		handler.ServeHTTP(w, req)

		require.False(t, called)
		require.Equal(t, http.StatusForbidden, w.Code)
		require.Contains(t, w.Body.String(), "enterprise access required")
	})

	t.Run("with empty enterprise ID", func(t *testing.T) {
		claims := appauth.Claims{
			Subject:      "user-123",
			Email:        "test@example.com",
			EnterpriseID: "",
			Name:         "Test User",
		}
		ctx := WithClaims(context.Background(), claims)

		req := httptest.NewRequest("GET", "/test", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		called := false
		handler := RequireEnterprise(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		handler.ServeHTTP(w, req)

		require.False(t, called)
		require.Equal(t, http.StatusForbidden, w.Code)
		require.Contains(t, w.Body.String(), "enterprise access required")
	})
}

func TestEnterpriseID(t *testing.T) {
	t.Run("with enterprise ID in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), enterpriseKey, "ent-123")
		enterpriseID := EnterpriseID(ctx)
		require.Equal(t, "ent-123", enterpriseID)
	})

	t.Run("without enterprise ID in context", func(t *testing.T) {
		ctx := context.Background()
		enterpriseID := EnterpriseID(ctx)
		require.Equal(t, "", enterpriseID)
	})

	t.Run("with wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), enterpriseKey, 123)
		enterpriseID := EnterpriseID(ctx)
		require.Equal(t, "", enterpriseID)
	})
}

func TestWithClaims(t *testing.T) {
	t.Run("adds claims to context", func(t *testing.T) {
		claims := appauth.Claims{
			Subject:      "user-123",
			Email:        "test@example.com",
			EnterpriseID: "ent-1",
			Name:         "Test User",
		}

		ctx := context.Background()
		ctxWithClaims := WithClaims(ctx, claims)

		retrievedClaims, ok := ClaimsFromCtx(ctxWithClaims)
		require.True(t, ok)
		require.Equal(t, "user-123", retrievedClaims.Subject)
		require.Equal(t, "test@example.com", retrievedClaims.Email)
		require.Equal(t, "ent-1", retrievedClaims.EnterpriseID)
		require.Equal(t, "Test User", retrievedClaims.Name)
	})
}
