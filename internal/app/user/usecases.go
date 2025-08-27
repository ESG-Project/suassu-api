package user

import (
	"context"
	"errors"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	repo   Repo
	hasher Hasher
}

func NewService(r Repo, h Hasher) *Service { return &Service{repo: r, hasher: h} }

type CreateInput struct {
	Name         string
	Email        string
	Password     string
	Document     string
	Phone        *string
	AddressID    *string
	RoleID       *string
	EnterpriseID string
}

func (s *Service) Create(ctx context.Context, enterpriseID string, in CreateInput) (string, error) {
	if in.Name == "" || in.Email == "" || in.Password == "" || in.Document == "" || enterpriseID == "" {
		return "", apperr.New(apperr.CodeInvalid, "missing required fields")
	}
	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return "", err
	}

	id := uuid.NewString()
	user := domainuser.NewUser(id, in.Name, in.Email, hash, in.Document, enterpriseID)

	// Set optional fields
	if in.Phone != nil {
		user.SetPhone(in.Phone)
	}
	if in.AddressID != nil {
		user.SetAddressID(in.AddressID)
	}
	if in.RoleID != nil {
		user.SetRoleID(in.RoleID)
	}

	// Validate user before saving
	if err := user.Validate(); err != nil {
		return "", apperr.Wrap(err, apperr.CodeInvalid, "invalid user data")
	}

	err = s.repo.Create(ctx, user)
	return id, err
}

func (s *Service) List(ctx context.Context, enterpriseID string, limit, offset int32) ([]domainuser.User, error) {
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	users, err := s.repo.List(ctx, enterpriseID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Converter ponteiros para valores
	result := make([]domainuser.User, len(users))
	for i, user := range users {
		result[i] = *user
	}
	return result, nil
}

func (s *Service) GetByEmailInTenant(ctx context.Context, enterpriseID string, email string) (*domainuser.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	return s.repo.GetByEmailInTenant(ctx, enterpriseID, email)
}
