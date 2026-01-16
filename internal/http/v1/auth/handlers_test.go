package authhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/http/cookie"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	authhttp "github.com/ESG-Project/suassu-api/internal/http/v1/auth"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

/*** FAKES ***/

type fakeAuthService struct {
	token               string
	refreshToken        string
	err                 error
	validateTokenResult bool
	validateTokenErr    error
	refreshErr          error
	logoutErr           error
}

func (f *fakeAuthService) SignIn(ctx context.Context, in appauth.SignInInput) (appauth.SignInOutput, error) {
	if f.err != nil {
		return appauth.SignInOutput{}, apperr.New(apperr.CodeUnauthorized, "invalid credentials")
	}
	refreshExpires := time.Now().Add(7 * 24 * time.Hour)
	return appauth.SignInOutput{
		AccessToken:      f.token,
		RefreshToken:     f.refreshToken,
		RefreshExpiresAt: &refreshExpires,
		ExpiresIn:        900, // 15 min
	}, nil
}

func (f *fakeAuthService) Refresh(ctx context.Context, in appauth.RefreshInput) (appauth.RefreshOutput, error) {
	if f.refreshErr != nil {
		return appauth.RefreshOutput{}, f.refreshErr
	}
	refreshExpires := time.Now().Add(7 * 24 * time.Hour)
	return appauth.RefreshOutput{
		AccessToken:      "new-access-token",
		RefreshToken:     "new-refresh-token",
		RefreshExpiresAt: &refreshExpires,
		ExpiresIn:        900,
	}, nil
}

func (f *fakeAuthService) Logout(ctx context.Context, userID string) error {
	return f.logoutErr
}

func (f *fakeAuthService) GetMe(ctx context.Context, userID string, enterpriseID string) (*types.UserWithDetails, error) {
	return &types.UserWithDetails{ID: userID, Name: "Ana", Email: "ana@ex.com", EnterpriseID: enterpriseID}, nil
}

func (f *fakeAuthService) GetMyPermissions(ctx context.Context, userID string, enterpriseID string) (*types.UserPermissions, error) {
	return &types.UserPermissions{ID: userID, Name: "Ana", RoleTitle: "Admin", Permissions: nil}, nil
}

func (f *fakeAuthService) ValidateToken(ctx context.Context, token string) (bool, error) {
	if f.validateTokenErr != nil {
		return false, f.validateTokenErr
	}
	return f.validateTokenResult, nil
}

// fakeTokenIssuer implementa appauth.TokenIssuer.
// Aqui corrigimos a assinatura de NewAccessToken para aceitar domain.User.
type fakeTokenIssuer struct {
	expect string
	claims appauth.Claims
	err    error
}

func (f *fakeTokenIssuer) NewAccessToken(_ *domainuser.User) (string, error) {
	// não usado nos testes deste arquivo
	return "unused", nil
}

func (f *fakeTokenIssuer) NewRefreshToken() (token string, hash string, expiresAt time.Time, err error) {
	return "refresh-token", "hash", time.Now().Add(7 * 24 * time.Hour), nil
}

func (f *fakeTokenIssuer) Parse(token string) (appauth.Claims, error) {
	if f.err != nil {
		return appauth.Claims{}, f.err
	}
	if token != f.expect {
		return appauth.Claims{}, errors.New("invalid token")
	}
	return f.claims, nil
}

func (f *fakeTokenIssuer) GetAccessTTL() int64 {
	return 900 // 15 min
}

func (f *fakeTokenIssuer) GetRefreshTTL() time.Duration {
	return 7 * 24 * time.Hour
}

/*** HELPERS ***/

func newTestCookieManager() *cookie.Manager {
	return cookie.NewManager(cookie.Config{
		Domain: "",
		Secure: false,
	})
}

func newAuthRouterPublic(svc *fakeAuthService) http.Handler {
	h := authhttp.NewHandler(svc, newTestCookieManager())
	r := chi.NewRouter()
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Route("/auth", func(auth chi.Router) {
			auth.Group(func(pub chi.Router) {
				h.RegisterPublic(pub) // POST /api/v1/auth/login
			})
		})
	})
	return r
}

func newAuthRouterPrivate(tokenIssuer *fakeTokenIssuer) http.Handler {
	h := authhttp.NewHandler(&fakeAuthService{}, newTestCookieManager())
	r := chi.NewRouter()
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Route("/auth", func(auth chi.Router) {
			auth.Group(func(priv chi.Router) {
				priv.Use(httpmw.AuthJWT(tokenIssuer))
				priv.Use(httpmw.RequireEnterprise)
				h.RegisterPrivate(priv) // GET /api/v1/auth/me
			})
		})
	})
	return r
}

/*** TESTES ***/

func TestAuth_Login_OK(t *testing.T) {
	svc := &fakeAuthService{token: "token-ok"}
	router := newAuthRouterPublic(svc)

	body := map[string]any{
		"email":    "ana@ex.com",
		"password": "123",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	var resp struct {
		AccessToken string `json:"accessToken"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "token-ok", resp.AccessToken)
}

func TestAuth_Login_InvalidBody(t *testing.T) {
	svc := &fakeAuthService{token: "token-ok"}
	router := newAuthRouterPublic(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestAuth_Login_InvalidCredentials(t *testing.T) {
	svc := &fakeAuthService{err: errors.New("invalid credentials")}
	router := newAuthRouterPublic(svc)

	body := map[string]any{
		"email":    "ana@ex.com",
		"password": "wrong",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_Me_OK(t *testing.T) {
	issuer := &fakeTokenIssuer{
		expect: "good-token",
		claims: appauth.Claims{
			Subject:      "u1",
			Email:        "ana@ex.com",
			Name:         "Ana",
			EnterpriseID: "ent-1",
			RoleID:       nil,
		},
	}
	router := newAuthRouterPrivate(issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "u1", resp["id"])              // MeOut.id
	require.Equal(t, "ana@ex.com", resp["email"])   // MeOut.email
	require.Equal(t, "Ana", resp["name"])           // MeOut.name
	require.Equal(t, "ent-1", resp["enterpriseId"]) // MeOut.enterpriseId
}

func TestAuth_Me_Unauthorized_NoHeader(t *testing.T) {
	issuer := &fakeTokenIssuer{
		expect: "good-token",
		claims: appauth.Claims{Subject: "u1"},
	}
	router := newAuthRouterPrivate(issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_Me_Unauthorized_BadToken(t *testing.T) {
	issuer := &fakeTokenIssuer{
		expect: "good-token",
		claims: appauth.Claims{Subject: "u1"},
	}
	router := newAuthRouterPrivate(issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_ValidateToken_OK(t *testing.T) {
	svc := &fakeAuthService{
		validateTokenResult: true,
	}
	router := newAuthRouterPublic(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate-token", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	require.Empty(t, rr.Body.String())
}

func TestAuth_ValidateToken_InvalidToken(t *testing.T) {
	svc := &fakeAuthService{
		validateTokenResult: false,
	}
	router := newAuthRouterPublic(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate-token", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]bool
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.False(t, resp["valid"])
}

func TestAuth_ValidateToken_EmptyToken(t *testing.T) {
	svc := &fakeAuthService{}
	router := newAuthRouterPublic(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate-token", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestAuth_ValidateToken_WithoutBearerPrefix(t *testing.T) {
	svc := &fakeAuthService{
		validateTokenResult: true,
	}
	router := newAuthRouterPublic(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate-token", nil)
	req.Header.Set("Authorization", "valid-token")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	require.Empty(t, rr.Body.String())
}

func TestAuth_Refresh_OK(t *testing.T) {
	svc := &fakeAuthService{}
	router := newAuthRouterPublic(svc)

	body := map[string]any{
		"refreshToken": "valid-refresh-token",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expiresIn"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "new-access-token", resp.AccessToken)
	require.Equal(t, 900, resp.ExpiresIn)

	// Verificar que o refresh token foi setado como cookie HttpOnly
	cookies := rr.Result().Cookies()
	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}
	require.NotNil(t, refreshCookie, "refresh_token cookie should be set")
	require.Equal(t, "new-refresh-token", refreshCookie.Value)
	require.True(t, refreshCookie.HttpOnly, "cookie should be HttpOnly")
}

func TestAuth_Refresh_NoToken(t *testing.T) {
	svc := &fakeAuthService{}
	router := newAuthRouterPublic(svc)

	// Sem cookie e sem body - deve retornar 401
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_Refresh_FromCookie(t *testing.T) {
	svc := &fakeAuthService{}
	router := newAuthRouterPublic(svc)

	// Token vem do cookie HttpOnly
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "valid-refresh-token",
	})
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expiresIn"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "new-access-token", resp.AccessToken)
}

func TestAuth_Refresh_InvalidToken(t *testing.T) {
	svc := &fakeAuthService{
		refreshErr: apperr.New(apperr.CodeUnauthorized, "invalid refresh token"),
	}
	router := newAuthRouterPublic(svc)

	body := map[string]any{
		"refreshToken": "invalid-token",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_Logout_OK(t *testing.T) {
	issuer := &fakeTokenIssuer{
		expect: "good-token",
		claims: appauth.Claims{
			Subject:      "u1",
			Email:        "ana@ex.com",
			Name:         "Ana",
			EnterpriseID: "ent-1",
			RoleID:       nil,
		},
	}
	router := newAuthRouterPrivate(issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
}

func TestAuth_Logout_Unauthorized(t *testing.T) {
	issuer := &fakeTokenIssuer{
		expect: "good-token",
		claims: appauth.Claims{Subject: "u1"},
	}
	router := newAuthRouterPrivate(issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_Login_WithRefreshToken(t *testing.T) {
	svc := &fakeAuthService{
		token:        "access-token",
		refreshToken: "refresh-token",
	}
	router := newAuthRouterPublic(svc)

	body := map[string]any{
		"email":    "ana@ex.com",
		"password": "123",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expiresIn"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "access-token", resp.AccessToken)
	require.Equal(t, 900, resp.ExpiresIn)

	// Verificar que o refresh token foi setado como cookie HttpOnly
	cookies := rr.Result().Cookies()
	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}
	require.NotNil(t, refreshCookie, "refresh_token cookie should be set")
	require.Equal(t, "refresh-token", refreshCookie.Value)
	require.True(t, refreshCookie.HttpOnly, "cookie should be HttpOnly")
}
