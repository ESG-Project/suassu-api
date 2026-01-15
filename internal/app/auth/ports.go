package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainrt "github.com/ESG-Project/suassu-api/internal/domain/refreshtoken"
	domain "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/google/uuid"
)

// TokenIssuer interface para geração de tokens JWT.
type TokenIssuer interface {
	NewAccessToken(u *domain.User) (string, error)
	NewRefreshToken() (token string, hash string, expiresAt time.Time, err error)
	Parse(token string) (Claims, error)
	GetAccessTTL() int64
	GetRefreshTTL() time.Duration
}

// RefreshTokenRepo interface para operações com refresh tokens.
type RefreshTokenRepo interface {
	Create(ctx context.Context, token *domainrt.RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (*domainrt.RefreshToken, error)
	GetTokenIfValid(ctx context.Context, tokenHash string) (*domainrt.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	RotateToken(ctx context.Context, oldTokenID string, newToken *domainrt.RefreshToken) error
}

// Claims representa os claims extraídos do JWT.
type Claims struct {
	Subject      string
	Email        string
	EnterpriseID string
	RoleID       *string
	Name         string
}

// Service representa o serviço de autenticação.
type Service struct {
	users         appuser.Repo
	userSvc       appuser.ServiceInterface
	hasher        appuser.Hasher
	tokens        TokenIssuer
	refreshTokens RefreshTokenRepo
}

// NewService cria um novo serviço de autenticação (sem refresh tokens).
func NewService(users appuser.Repo, userSvc appuser.ServiceInterface, hasher appuser.Hasher, tokens TokenIssuer) *Service {
	return &Service{users: users, userSvc: userSvc, hasher: hasher, tokens: tokens}
}

// NewServiceWithRefresh cria um novo serviço de autenticação com suporte a refresh tokens.
func NewServiceWithRefresh(users appuser.Repo, userSvc appuser.ServiceInterface, hasher appuser.Hasher, tokens TokenIssuer, refreshTokens RefreshTokenRepo) *Service {
	return &Service{users: users, userSvc: userSvc, hasher: hasher, tokens: tokens, refreshTokens: refreshTokens}
}

// SignInInput representa os dados de entrada para login.
type SignInInput struct {
	Email    string
	Password string
}

// SignInOutput representa os dados de saída do login.
type SignInOutput struct {
	AccessToken      string     `json:"accessToken"`
	RefreshToken     string     `json:"refreshToken,omitempty"`
	RefreshExpiresAt *time.Time `json:"-"` // não serializado, usado internamente para cookie
	ExpiresIn        int        `json:"expiresIn"` // segundos até o access token expirar
}

// RefreshInput representa os dados de entrada para refresh.
type RefreshInput struct {
	RefreshToken string
}

// RefreshOutput representa os dados de saída do refresh.
type RefreshOutput struct {
	AccessToken      string     `json:"accessToken"`
	RefreshToken     string     `json:"refreshToken"`
	RefreshExpiresAt *time.Time `json:"-"` // não serializado, usado internamente para cookie
	ExpiresIn        int        `json:"expiresIn"` // segundos até o access token expirar
}

// SignIn realiza o login do usuário.
func (s *Service) SignIn(ctx context.Context, in SignInInput) (SignInOutput, error) {
	u, err := s.users.GetByEmailForAuth(ctx, in.Email)
	if err != nil {
		return SignInOutput{}, apperr.Wrap(err, apperr.CodeNotFound, "user not found")
	}
	if err := s.hasher.Compare(u.PasswordHash, in.Password); err != nil {
		return SignInOutput{}, apperr.New(apperr.CodeUnauthorized, "invalid credentials")
	}

	// Gerar access token
	accessToken, err := s.tokens.NewAccessToken(u)
	if err != nil {
		return SignInOutput{}, apperr.Wrap(err, apperr.CodeInternal, "failed to generate access token")
	}

	output := SignInOutput{
		AccessToken: accessToken,
		ExpiresIn:   int(s.tokens.GetAccessTTL()),
	}

	// Gerar refresh token se o repositório estiver configurado
	if s.refreshTokens != nil {
		refreshToken, tokenHash, expiresAt, err := s.tokens.NewRefreshToken()
		if err != nil {
			return SignInOutput{}, apperr.Wrap(err, apperr.CodeInternal, "failed to generate refresh token")
		}

		// Salvar refresh token no banco
		rt := domainrt.NewRefreshToken(uuid.NewString(), u.ID, tokenHash, expiresAt)
		if err := s.refreshTokens.Create(ctx, rt); err != nil {
			return SignInOutput{}, apperr.Wrap(err, apperr.CodeInternal, "failed to save refresh token")
		}

		output.RefreshToken = refreshToken
		output.RefreshExpiresAt = &expiresAt
	}

	return output, nil
}

// Refresh renova os tokens usando um refresh token válido.
func (s *Service) Refresh(ctx context.Context, in RefreshInput) (RefreshOutput, error) {
	if s.refreshTokens == nil {
		return RefreshOutput{}, apperr.New(apperr.CodeInternal, "refresh tokens not configured")
	}

	// Calcular hash do refresh token recebido
	tokenHash := hashRefreshToken(in.RefreshToken)

	// Buscar e validar o refresh token
	oldToken, err := s.refreshTokens.GetTokenIfValid(ctx, tokenHash)
	if err != nil {
		return RefreshOutput{}, apperr.Wrap(err, apperr.CodeUnauthorized, "invalid refresh token")
	}

	// Buscar usuário
	u, err := s.users.GetByIDForRefresh(ctx, oldToken.UserID)
	if err != nil {
		return RefreshOutput{}, apperr.Wrap(err, apperr.CodeNotFound, "user not found")
	}

	// Gerar novo access token
	accessToken, err := s.tokens.NewAccessToken(u)
	if err != nil {
		return RefreshOutput{}, apperr.Wrap(err, apperr.CodeInternal, "failed to generate access token")
	}

	// Gerar novo refresh token (rotação)
	newRefreshToken, newTokenHash, expiresAt, err := s.tokens.NewRefreshToken()
	if err != nil {
		return RefreshOutput{}, apperr.Wrap(err, apperr.CodeInternal, "failed to generate refresh token")
	}

	// Rotacionar tokens (revogar antigo, criar novo)
	newRT := domainrt.NewRefreshToken(uuid.NewString(), u.ID, newTokenHash, expiresAt)
	if err := s.refreshTokens.RotateToken(ctx, oldToken.ID, newRT); err != nil {
		return RefreshOutput{}, apperr.Wrap(err, apperr.CodeInternal, "failed to rotate refresh token")
	}

	return RefreshOutput{
		AccessToken:      accessToken,
		RefreshToken:     newRefreshToken,
		RefreshExpiresAt: &expiresAt,
		ExpiresIn:        int(s.tokens.GetAccessTTL()),
	}, nil
}

// Logout revoga todos os refresh tokens do usuário.
func (s *Service) Logout(ctx context.Context, userID string) error {
	if s.refreshTokens == nil {
		return nil // Nada a fazer se refresh tokens não estão configurados
	}
	return s.refreshTokens.RevokeAllForUser(ctx, userID)
}

// GetMe retorna as informações do usuário logado.
func (s *Service) GetMe(ctx context.Context, userID string, enterpriseID string) (*types.UserWithDetails, error) {
	return s.userSvc.GetUserWithDetails(ctx, userID, enterpriseID)
}

// GetMyPermissions retorna as permissões do usuário logado.
func (s *Service) GetMyPermissions(ctx context.Context, userID string, enterpriseID string) (*types.UserPermissions, error) {
	return s.userSvc.GetUserPermissionsWithRole(ctx, userID, enterpriseID)
}

// ValidateToken valida um token JWT.
func (s *Service) ValidateToken(ctx context.Context, token string) (bool, error) {
	_, err := s.tokens.Parse(token)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// hashRefreshToken cria um hash SHA-256 de um refresh token.
func hashRefreshToken(token string) string {
	hashBytes := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hashBytes[:])
}
