package userhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
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
		return "", errors.New("missing required fields")
	}
	return f.createdID, nil
}

func (f *fakeSvc) List(ctx context.Context, enterpriseID string, limit, offset int32) ([]domainuser.User, error) {
	return f.users, nil
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

	req := httptest.NewRequest(http.MethodGet, "/users/?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	require.Equal(t, float64(10), response["limit"])
	require.Equal(t, float64(0), response["offset"])
	require.Equal(t, float64(2), response["count"])

	items := response["items"].([]any)
	require.Len(t, items, 2)

	// Verifica primeiro usuário
	firstUser := items[0].(map[string]any)
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

	// Verifica valores padrão
	require.Equal(t, float64(50), response["limit"])
	require.Equal(t, float64(0), response["offset"])
	require.Equal(t, float64(0), response["count"])
}

func TestPOST_Users_Create_InvalidBody(t *testing.T) {
	svc := &fakeSvc{createdID: "uuid-1"}
	router := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
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
