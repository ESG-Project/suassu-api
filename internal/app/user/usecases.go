package user

import (
	"context"
	"errors"

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

func (s *Service) Create(ctx context.Context, in CreateInput) (string, error) {
	if in.Name == "" || in.Email == "" || in.Password == "" || in.Document == "" || in.EnterpriseID == "" {
		return "", errors.New("missing required fields")
	}
	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return "", err
	}

	id := uuid.NewString()
	user := domainuser.NewUser(id, in.Name, in.Email, hash, in.Document, in.EnterpriseID)

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
		return "", err
	}

	err = s.repo.Create(ctx, user)
	return id, err
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	return s.repo.GetByEmail(ctx, email)
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]*domainuser.User, error) {
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}
