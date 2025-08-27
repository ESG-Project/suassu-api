package authhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	authhttp "github.com/ESG-Project/suassu-api/internal/http/v1/auth"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

/*** FAKES ***/

type fakeAuthService struct {
	token string
	err   error
}

func (f *fakeAuthService) SignIn(ctx context.Context, in appauth.SignInInput) (appauth.SignInOutput, error) {
	if f.err != nil {
		return appauth.SignInOutput{}, apperr.New(apperr.CodeUnauthorized, "invalid credentials")
	}
	return appauth.SignInOutput{AccessToken: f.token}, nil
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
func (f *fakeTokenIssuer) Parse(token string) (appauth.Claims, error) {
	if f.err != nil {
		return appauth.Claims{}, f.err
	}
	if token != f.expect {
		return appauth.Claims{}, errors.New("invalid token")
	}
	return f.claims, nil
}

/*** HELPERS ***/

func newAuthRouterPublic(svc *fakeAuthService) http.Handler {
	h := authhttp.NewHandler(svc)
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
	h := authhttp.NewHandler(&fakeAuthService{})
	r := chi.NewRouter()
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Route("/auth", func(auth chi.Router) {
			auth.Group(func(priv chi.Router) {
				priv.Use(httpmw.AuthJWT(tokenIssuer))
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
	require.Equal(t, "u1", resp["sub"])
	require.Equal(t, "ana@ex.com", resp["email"])
	require.Equal(t, "Ana", resp["name"])
	require.Equal(t, "ent-1", resp["enterpriseId"])
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
