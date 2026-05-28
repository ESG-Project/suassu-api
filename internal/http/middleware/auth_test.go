package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// Mock TokenIssuer para teste
type mockTokenIssuer struct {
	shouldFail bool
	claims     appauth.Claims
}

func (m *mockTokenIssuer) NewAccessToken(u *user.User) (string, error) {
	if m.shouldFail {
		return "", errors.New("token generation failed")
	}
	return "mock-token-123", nil
}

func (m *mockTokenIssuer) NewRefreshToken() (token string, hash string, expiresAt time.Time, err error) {
	if m.shouldFail {
		return "", "", time.Time{}, errors.New("refresh token generation failed")
	}
	return "mock-refresh-token", "mock-hash", time.Now().Add(7 * 24 * time.Hour), nil
}

func (m *mockTokenIssuer) Parse(token string) (appauth.Claims, error) {
	if m.shouldFail {
		return appauth.Claims{}, errors.New("token parsing failed")
	}
	return m.claims, nil
}

func (m *mockTokenIssuer) GetAccessTTL() int64 {
	return 900 // 15 min
}

func (m *mockTokenIssuer) GetRefreshTTL() time.Duration {
	return 7 * 24 * time.Hour
}

// Handler protegido para teste
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromCtx(r.Context())
	if !ok {
		http.Error(w, "claims not found", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"subject":      claims.Subject,
		"email":        claims.Email,
		"enterpriseId": claims.EnterpriseID,
		"name":         claims.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Teste simples para verificar se o middleware está funcionando
func TestAuthJWT_Simple(t *testing.T) {
	// Claims de teste
	testClaims := appauth.Claims{
		Subject:      "user-123",
		Email:        "teste@exemplo.com",
		EnterpriseID: "empresa-teste",
		Name:         "Usuário Teste",
	}

	// Mock TokenIssuer que sempre retorna sucesso
	mockIssuer := &mockTokenIssuer{
		shouldFail: false,
		claims:     testClaims,
	}

	// Setup router com middleware
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(middleware.AuthJWT(mockIssuer))
		r.Get("/private/ping", protectedHandler)
	})

	// Teste com token válido
	t.Run("Bearer válido", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/private/ping", nil)
		req.Header.Set("Authorization", "Bearer mock-token-123")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		require.Equal(t, "user-123", response["subject"])
		require.Equal(t, "teste@exemplo.com", response["email"])
	})

	// Teste sem token
	t.Run("Sem Authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/private/ping", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.Contains(t, w.Body.String(), "bearer token required")
	})

	// Teste com token inválido
	t.Run("Token inválido", func(t *testing.T) {
		// Mock que falha no parsing
		failingIssuer := &mockTokenIssuer{
			shouldFail: true,
			claims:     testClaims,
		}

		router := chi.NewRouter()
		router.Group(func(r chi.Router) {
			r.Use(middleware.AuthJWT(failingIssuer))
			r.Get("/private/ping", protectedHandler)
		})

		req := httptest.NewRequest("GET", "/private/ping", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.Contains(t, w.Body.String(), "invalid token")
	})
}

func TestClaimsFromCtx(t *testing.T) {
	t.Run("Claims não encontradas no contexto", func(t *testing.T) {
		ctx := context.Background()
		claims, ok := middleware.ClaimsFromCtx(ctx)

		require.False(t, ok)
		require.Equal(t, appauth.Claims{}, claims)
	})
}
