package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	domain "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/stretchr/testify/require"
)

// Fakes para teste
type fakeRepo struct {
	users map[string]*domain.User
	err   error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		users: map[string]*domain.User{
			"teste@exemplo.com": domain.NewUser(
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

func (f *fakeRepo) Create(ctx context.Context, u *domain.User) error {
	f.users[u.Email] = u
	return nil
}

func (f *fakeRepo) List(ctx context.Context, enterpriseID string, limit int32, after *domain.UserCursorKey) ([]*domain.User, domain.PageInfo, error) {
	if f.err != nil {
		return nil, domain.PageInfo{}, f.err
	}
	users := []*domain.User{
		{ID: "user-1", Name: "Ana", Email: "ana@ex.com", EnterpriseID: "ent-1"},
		{ID: "user-2", Name: "Bob", Email: "bob@ex.com", EnterpriseID: "ent-1"},
	}
	return users, domain.PageInfo{}, nil
}

func (f *fakeRepo) GetByEmailForAuth(ctx context.Context, email string) (*domain.User, error) {
	if u, exists := f.users[email]; exists {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (f *fakeRepo) GetUserPermissionsWithRole(ctx context.Context, userID string, enterpriseID string) (*types.UserPermissions, error) {
	return &types.UserPermissions{ID: userID, Name: "Ana", RoleTitle: "Admin"}, nil
}

func (f *fakeRepo) GetUserWithDetails(ctx context.Context, userID string, enterpriseID string) (*types.UserWithDetails, error) {
	return &types.UserWithDetails{ID: userID, Name: "Ana", Email: "ana@ex.com", EnterpriseID: enterpriseID}, nil
}

type fakeHasher struct{}

func (f *fakeHasher) Hash(pw string) (string, error) {
	return "HASH_" + pw, nil
}

func (f *fakeHasher) Compare(hash, plain string) error {
	expectedHash := "HASH_" + plain
	if hash == expectedHash {
		return nil
	}
	return errors.New("password mismatch")
}

type fakeUserService struct{}

func (f *fakeUserService) Create(ctx context.Context, enterpriseID string, in appuser.CreateInput) (string, error) {
	return "", errors.New("not implemented in fake")
}

func (f *fakeUserService) List(ctx context.Context, enterpriseID string, limit int32, after *domain.UserCursorKey) ([]domain.User, *domain.PageInfo, error) {
	return nil, nil, errors.New("not implemented in fake")
}

func (f *fakeUserService) GetUserWithDetails(ctx context.Context, userID string, enterpriseID string) (*types.UserWithDetails, error) {
	return nil, errors.New("not implemented in fake")
}

func (f *fakeUserService) GetUserPermissionsWithRole(ctx context.Context, userID string, enterpriseID string) (*types.UserPermissions, error) {
	return &types.UserPermissions{ID: userID, Name: "Ana", RoleTitle: "Admin"}, nil
}

type fakeTokenIssuer struct {
	shouldFail bool
}

func (f *fakeTokenIssuer) NewAccessToken(u *domain.User) (string, error) {
	if f.shouldFail {
		return "", errors.New("token generation failed")
	}
	return "token-ok", nil
}

func (f *fakeTokenIssuer) Parse(token string) (appauth.Claims, error) {
	return appauth.Claims{}, errors.New("not implemented in fake")
}

func TestService_SignIn(t *testing.T) {
	tests := []struct {
		name        string
		input       appauth.SignInInput
		setupRepo   func() *fakeRepo
		setupHasher func() *fakeHasher
		setupTokens func() *fakeTokenIssuer
		wantToken   string
		wantErr     bool
	}{
		{
			name: "OK - credenciais válidas",
			input: appauth.SignInInput{
				Email:    "teste@exemplo.com",
				Password: "senha123",
			},
			setupRepo:   newFakeRepo,
			setupHasher: func() *fakeHasher { return &fakeHasher{} },
			setupTokens: func() *fakeTokenIssuer { return &fakeTokenIssuer{shouldFail: false} },
			wantToken:   "token-ok",
			wantErr:     false,
		},
		{
			name: "Email não encontrado",
			input: appauth.SignInInput{
				Email:    "inexistente@exemplo.com",
				Password: "senha123",
			},
			setupRepo:   newFakeRepo,
			setupHasher: func() *fakeHasher { return &fakeHasher{} },
			setupTokens: func() *fakeTokenIssuer { return &fakeTokenIssuer{shouldFail: false} },
			wantToken:   "",
			wantErr:     true,
		},
		{
			name: "Senha incorreta",
			input: appauth.SignInInput{
				Email:    "teste@exemplo.com",
				Password: "senha_errada",
			},
			setupRepo:   newFakeRepo,
			setupHasher: func() *fakeHasher { return &fakeHasher{} },
			setupTokens: func() *fakeTokenIssuer { return &fakeTokenIssuer{shouldFail: false} },
			wantToken:   "",
			wantErr:     true,
		},
		{
			name: "Falha ao emitir token",
			input: appauth.SignInInput{
				Email:    "teste@exemplo.com",
				Password: "senha123",
			},
			setupRepo:   newFakeRepo,
			setupHasher: func() *fakeHasher { return &fakeHasher{} },
			setupTokens: func() *fakeTokenIssuer { return &fakeTokenIssuer{shouldFail: true} },
			wantToken:   "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Setup
			repo := tt.setupRepo()
			hasher := tt.setupHasher()
			tokens := tt.setupTokens()

			// Criar um fake user service
			fakeUserSvc := &fakeUserService{}
			svc := appauth.NewService(repo, fakeUserSvc, hasher, tokens)

			// Execute
			result, err := svc.SignIn(ctx, tt.input)

			// Assertions
			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, result.AccessToken)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantToken, result.AccessToken)
			}
		})
	}
}
