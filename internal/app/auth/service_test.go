package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/stretchr/testify/require"
)

// Fakes para teste
type fakeRepo struct {
	users map[string]*user.User
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
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

func (f *fakeRepo) Create(ctx context.Context, u *user.User) error {
	f.users[u.Email] = u
	return nil
}

func (f *fakeRepo) List(ctx context.Context, enterpriseID string, limit, offset int32) ([]*user.User, error) {
	users := make([]*user.User, 0, len(f.users))
	for _, u := range f.users {
		users = append(users, u)
	}
	return users, nil
}

func (f *fakeRepo) GetByEmailForAuth(ctx context.Context, email string) (*user.User, error) {
	if u, exists := f.users[email]; exists {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (f *fakeRepo) GetByEmailInTenant(ctx context.Context, enterpriseID string, email string) (*user.User, error) {
	if u, exists := f.users[email]; exists {
		return u, nil
	}
	return nil, errors.New("user not found")
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

type fakeTokenIssuer struct {
	shouldFail bool
}

func (f *fakeTokenIssuer) NewAccessToken(u *user.User) (string, error) {
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

			svc := appauth.NewService(repo, hasher, tokens)

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
