package auth

import (
	"context"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domain "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/http/dto"
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
	users   appuser.Repo
	userSvc appuser.ServiceInterface
	hasher  appuser.Hasher
	tokens  TokenIssuer
}

func NewService(users appuser.Repo, userSvc appuser.ServiceInterface, hasher appuser.Hasher, tokens TokenIssuer) *Service {
	return &Service{users: users, userSvc: userSvc, hasher: hasher, tokens: tokens}
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
		return SignInOutput{}, apperr.Wrap(err, apperr.CodeNotFound, "user not found")
	}
	if err := s.hasher.Compare(u.PasswordHash, in.Password); err != nil {
		return SignInOutput{}, apperr.New(apperr.CodeUnauthorized, "invalid credentials")
	}
	tok, err := s.tokens.NewAccessToken(u)
	if err != nil {
		return SignInOutput{}, apperr.Wrap(err, apperr.CodeInternal, "failed to generate token")
	}
	return SignInOutput{AccessToken: tok}, nil
}

func (s *Service) GetMyPermissions(ctx context.Context, userID string, enterpriseID string) (*dto.MyPermissionsOut, error) {
	// Delega para o user service
	return s.userSvc.GetUserPermissionsWithRole(ctx, userID, enterpriseID)
}
