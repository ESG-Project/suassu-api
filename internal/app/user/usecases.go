package user

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/address"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/http/dto"
	postgres "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
	"github.com/google/uuid"
)

type Service struct {
	repo           Repo
	addressService *address.Service
	hasher         Hasher
	txm            postgres.TxManagerInterface
}

func NewService(r Repo, as *address.Service, h Hasher) *Service {
	return NewServiceWithTx(r, as, h, nil)
}

func NewServiceWithTx(r Repo, as *address.Service, h Hasher, txm postgres.TxManagerInterface) *Service {
	return &Service{repo: r, addressService: as, hasher: h, txm: txm}
}

type CreateInput struct {
	Name         string
	Email        string
	Password     string
	Document     string
	Phone        *string
	AddressID    *string
	Address      *address.CreateInput
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

	// Quando houver criação de endereço + usuário e TxManager disponível, usar transação
	if in.Address != nil {
		if s.txm == nil {
			return "", apperr.New(apperr.CodeInvalid, "tx manager required for address creation")
		}
		var createdID string
		err := s.txm.RunInTx(ctx, func(r postgres.Repos) error {
			addrSvc := address.NewService(r.Addresses(), s.hasher)
			addressID, err := addrSvc.HandleAddress(ctx, in.Address)
			if err != nil {
				return err
			}
			user.SetAddressID(&addressID)

			if err := r.Users().Create(ctx, user); err != nil {
				return err
			}
			createdID = id
			return nil
		})
		return createdID, err
	}

	if in.AddressID != nil {
		user.SetAddressID(in.AddressID)
	}
	err = s.repo.Create(ctx, user)

	return id, err
}

func (s *Service) GetByEmailInTenant(ctx context.Context, enterpriseID string, email string) (*domainuser.User, error) {
	if email == "" {
		return nil, apperr.New(apperr.CodeInvalid, "email is required")
	}
	return s.repo.GetByEmailInTenant(ctx, enterpriseID, email)
}

func (s *Service) List(ctx context.Context, enterpriseID string, limit int32, after *domainuser.UserCursorKey) ([]domainuser.User, *domainuser.PageInfo, error) {
	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	users, _, err := s.repo.List(ctx, enterpriseID, limit, after)
	if err != nil {
		return nil, nil, err
	}

	users, pageInfo := PaginateResult(users, limit)

	// Converter ponteiros para valores
	result := make([]domainuser.User, len(users))
	for i, user := range users {
		result[i] = *user
	}
	return result, &pageInfo, nil
}

func (s *Service) GetUserPermissionsWithRole(ctx context.Context, userID string, enterpriseID string) (*dto.MyPermissionsOut, error) {
	userWithPermissions, err := s.repo.GetUserPermissionsWithRole(ctx, userID, enterpriseID)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "user not found")
	}

	return userWithPermissions, nil
}

func (s *Service) GetUserWithDetails(ctx context.Context, userID string, enterpriseID string) (*dto.MeOut, error) {
	// Buscar usuário com todas as informações (incluindo endereço e empresa)
	userWithDetails, err := s.repo.GetUserWithDetails(ctx, userID, enterpriseID)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "user not found")
	}

	return userWithDetails, nil
}
