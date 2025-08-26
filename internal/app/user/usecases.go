package user

import (
	"context"
	"errors"

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
	err = s.repo.Create(ctx, Entity{
		ID:           id,
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: hash,
		Document:     in.Document,
		Phone:        in.Phone,
		AddressID:    in.AddressID,
		RoleID:       in.RoleID,
		EnterpriseID: in.EnterpriseID,
	})
	return id, err
}

func (s *Service) GetByEmail(ctx context.Context, email string) (Entity, error) {
	if email == "" {
		return Entity{}, errors.New("email is required")
	}
	return s.repo.GetByEmail(ctx, email)
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]Entity, error) {
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}
