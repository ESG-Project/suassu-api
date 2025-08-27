package auth

import (
	"context"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	domain "github.com/ESG-Project/suassu-api/internal/domain/user"
)

type TokenIssuer interface {
	NewAccessToken(u *domain.User) (string, error)
	Parse(token string) (Claims, error)
}

type Claims struct {
	Subject      string
	Email        string
	EnterpriseID string
	RoleID       *string
	Name         string
}

type Service struct {
	users  appuser.Repo
	hasher appuser.Hasher
	tokens TokenIssuer
}

func NewService(users appuser.Repo, hasher appuser.Hasher, tokens TokenIssuer) *Service {
	return &Service{users: users, hasher: hasher, tokens: tokens}
}

type SignInInput struct {
	Email    string
	Password string
}

type SignInOutput struct {
	AccessToken string `json:"accessToken"`
}

func (s *Service) SignIn(ctx context.Context, in SignInInput) (SignInOutput, error) {
	u, err := s.users.GetByEmailForAuth(ctx, in.Email)
	if err != nil {
		return SignInOutput{}, err
	}
	if err := s.hasher.Compare(u.PasswordHash, in.Password); err != nil {
		return SignInOutput{}, err
	}
	tok, err := s.tokens.NewAccessToken(u)
	if err != nil {
		return SignInOutput{}, err
	}
	return SignInOutput{AccessToken: tok}, nil
}
