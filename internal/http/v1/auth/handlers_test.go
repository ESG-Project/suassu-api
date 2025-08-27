package authhttp_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/domain/user"
	authhttp "github.com/ESG-Project/suassu-api/internal/http/v1/auth"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// Mock service para teste
type mockAuthService struct {
	shouldFail bool
	failReason string
}

func (m *mockAuthService) SignIn(ctx context.Context, in appauth.SignInInput) (appauth.SignInOutput, error) {
	if m.shouldFail {
		return appauth.SignInOutput{}, errors.New(m.failReason)
	}
	return appauth.SignInOutput{AccessToken: "fake-token-123"}, nil
}

// Mock repo e hasher para criar o service real
type mockUserRepo struct {
	users map[string]*user.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: map[string]*user.User{
			"teste@exemplo.com": user.NewUser(
				"user-123",
				"Usuário Teste",
				"teste@exemplo.com",
				"HASH_senha123",
				"12345678901",
				"empresa-teste",
			),
		},
	}
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	if u, exists := m.users[email]; exists {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepo) List(ctx context.Context, limit, offset int32) ([]*user.User, error) {
	users := make([]*user.User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}

type mockHasher struct{}

func (m *mockHasher) Hash(pw string) (string, error) {
	return "HASH_" + pw, nil
}

func (m *mockHasher) Compare(hash, plain string) error {
	expectedHash := "HASH_" + plain
	if hash == expectedHash {
		return nil
	}
	return errors.New("password mismatch")
}

type mockTokenIssuer struct {
	shouldFail bool
}

func (m *mockTokenIssuer) NewAccessToken(u *user.User) (string, error) {
	if m.shouldFail {
		return "", errors.New("token generation failed")
	}
	return "fake-token-123", nil
}

func (m *mockTokenIssuer) Parse(token string) (appauth.Claims, error) {
	return appauth.Claims{}, errors.New("not implemented in mock")
}

func TestAuthRoutes_Login(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupService   func() *appauth.Service
		expectedStatus int
		expectedBody   map[string]interface{}
		checkHeaders   func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "OK - credenciais válidas",
			body: `{"email":"teste@exemplo.com","password":"senha123"}`,
			setupService: func() *appauth.Service {
				repo := newMockUserRepo()
				hasher := &mockHasher{}
				tokens := &mockTokenIssuer{shouldFail: false}
				return appauth.NewService(repo, hasher, tokens)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"accessToken": "fake-token-123",
			},
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, "application/json", w.Header().Get("Content-Type"))
			},
		},
		{
			name: "Body inválido - JSON quebrado",
			body: `{"email":"teste@exemplo.com","password":"senha123"`,
			setupService: func() *appauth.Service {
				repo := newMockUserRepo()
				hasher := &mockHasher{}
				tokens := &mockTokenIssuer{shouldFail: false}
				return appauth.NewService(repo, hasher, tokens)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
			checkHeaders:   func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name: "Credenciais inválidas - email não encontrado",
			body: `{"email":"inexistente@exemplo.com","password":"senha123"}`,
			setupService: func() *appauth.Service {
				repo := newMockUserRepo()
				hasher := &mockHasher{}
				tokens := &mockTokenIssuer{shouldFail: false}
				return appauth.NewService(repo, hasher, tokens)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   nil,
			checkHeaders:   func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name: "Credenciais inválidas - senha incorreta",
			body: `{"email":"teste@exemplo.com","password":"senha_errada"}`,
			setupService: func() *appauth.Service {
				repo := newMockUserRepo()
				hasher := &mockHasher{}
				tokens := &mockTokenIssuer{shouldFail: false}
				return appauth.NewService(repo, hasher, tokens)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   nil,
			checkHeaders:   func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			svc := tt.setupService()
			router := chi.NewRouter()
			router.Mount("/auth", authhttp.Routes(svc))

			// Create request
			req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assertions
			require.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					require.Equal(t, expectedValue, response[key])
				}
			}

			// Check headers
			tt.checkHeaders(t, w)

			// Verify no sensitive data is exposed
			bodyStr := w.Body.String()
			require.NotContains(t, bodyStr, "PasswordHash")
			require.NotContains(t, bodyStr, "password")
		})
	}
}

func TestAuthRoutes_Me(t *testing.T) {
	t.Run("Not implemented endpoint", func(t *testing.T) {
		// Setup
		repo := newMockUserRepo()
		hasher := &mockHasher{}
		tokens := &mockTokenIssuer{shouldFail: false}
		svc := appauth.NewService(repo, hasher, tokens)

		router := chi.NewRouter()
		router.Mount("/auth", authhttp.Routes(svc))

		// Create request
		req := httptest.NewRequest("GET", "/auth/me", nil)
		w := httptest.NewRecorder()

		// Execute
		router.ServeHTTP(w, req)

		// Assertions
		require.Equal(t, http.StatusNotImplemented, w.Code)
	})
}
