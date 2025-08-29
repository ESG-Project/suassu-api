package userhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	userhttp "github.com/ESG-Project/suassu-api/internal/http/v1/user"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type fakeSvc struct {
	createdID string
	users     []domainuser.User
}

func (f *fakeSvc) Create(ctx context.Context, enterpriseID string, in appuser.CreateInput) (string, error) {
	// Simula a validação do serviço real
	if in.Name == "" || in.Email == "" || in.Password == "" || in.Document == "" || enterpriseID == "" {
		return "", apperr.New(apperr.CodeInvalid, "missing required fields")
	}
	return f.createdID, nil
}

func (f *fakeSvc) List(ctx context.Context, enterpriseID string, limit int32, after *domainuser.UserCursorKey) ([]domainuser.User, *domainuser.PageInfo, error) {
	return f.users, &domainuser.PageInfo{HasMore: false, Next: nil}, nil
}

func (f *fakeSvc) ListAfter(ctx context.Context, enterpriseID string, limit int32, after *domainuser.UserCursorKey) ([]domainuser.User, *domainuser.PageInfo, error) {
	return f.users, &domainuser.PageInfo{HasMore: false}, nil
}

// Helper para criar router de teste com middlewares simulados
func newTestRouter(svc userhttp.Service) http.Handler {
	r := chi.NewRouter()

	// Simula o middleware AuthJWT
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simula claims válidos com enterpriseID
			claims := appauth.Claims{
				Subject:      "test-user",
				Email:        "test@example.com",
				Name:         "Test User",
				EnterpriseID: "test-enterprise",
				RoleID:       nil,
			}
			ctx := httpmw.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Simula o middleware RequireEnterprise
	r.Use(httpmw.RequireEnterprise)

	r.Mount("/users", userhttp.Routes(svc))
	return r
}

func TestPOST_Users_Create(t *testing.T) {
	svc := &fakeSvc{createdID: "uuid-1"}
	router := newTestRouter(svc)

	body := map[string]any{
		"name":     "Ana",
		"email":    "ana@ex.com",
		"password": "123",
		"document": "999",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.Contains(t, w.Body.String(), `"id":"uuid-1"`)
}

func TestGET_Users_List(t *testing.T) {
	testUsers := []domainuser.User{
		{
			ID:           "user-1",
			Name:         "Ana",
			Email:        "ana@ex.com",
			Document:     "12345678900",
			EnterpriseID: "test-enterprise",
		},
		{
			ID:           "user-2",
			Name:         "João",
			Email:        "joao@ex.com",
			Document:     "98765432100",
			EnterpriseID: "test-enterprise",
		},
	}

	svc := &fakeSvc{users: testUsers}
	router := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/users/?limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	// Verifica estrutura do envelope
	require.Contains(t, response, "data")
	require.Contains(t, response, "meta")

	data := response["data"].([]any)
	require.Len(t, data, 2)

	meta := response["meta"].(map[string]any)
	require.Equal(t, float64(10), meta["limit"])
	require.Equal(t, false, meta["hasMore"])

	// Verifica primeiro usuário
	firstUser := data[0].(map[string]any)
	require.Equal(t, "user-1", firstUser["id"])
	require.Equal(t, "Ana", firstUser["name"])
	require.Equal(t, "ana@ex.com", firstUser["email"])
	require.Equal(t, "test-enterprise", firstUser["enterpriseId"])
}

func TestGET_Users_List_DefaultParams(t *testing.T) {
	svc := &fakeSvc{users: []domainuser.User{}}
	router := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	// Verifica estrutura do envelope
	require.Contains(t, response, "data")
	require.Contains(t, response, "meta")

	data := response["data"].([]any)
	require.Len(t, data, 0)

	meta := response["meta"].(map[string]any)
	require.Equal(t, float64(50), meta["limit"])
	require.Equal(t, false, meta["hasMore"])
}

func TestPOST_Users_Create_InvalidBody(t *testing.T) {
	svc := &fakeSvc{createdID: "uuid-1"}
	router := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestPOST_Users_Create_MissingRequiredFields(t *testing.T) {
	svc := &fakeSvc{createdID: "uuid-1"}
	router := newTestRouter(svc)

	// Testa com campos obrigatórios faltando
	testCases := []map[string]any{
		{"email": "ana@ex.com", "password": "123", "document": "999"}, // sem name
		{"name": "Ana", "password": "123", "document": "999"},         // sem email
		{"name": "Ana", "email": "ana@ex.com", "document": "999"},     // sem password
		{"name": "Ana", "email": "ana@ex.com", "password": "123"},     // sem document
	}

	for _, body := range testCases {
		t.Run("missing_required_field", func(t *testing.T) {
			b, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// O handler deve retornar erro de validação
			require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		})
	}
}

func TestGET_Users_List_InvalidCursor(t *testing.T) {
	svc := &fakeSvc{users: []domainuser.User{}}
	router := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/users/?cursor=invalid-base64", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
	require.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.Contains(t, response, "error")
	require.Contains(t, response, "meta")
}
