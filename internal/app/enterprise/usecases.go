package enterprise

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/address"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainenterprise "github.com/ESG-Project/suassu-api/internal/domain/enterprise"
	"github.com/google/uuid"
)

type Service struct {
	repo           Repo
	addressService *address.Service
	hasher         Hasher
}

func NewService(r Repo, as *address.Service, h Hasher) *Service {
	return &Service{repo: r, addressService: as, hasher: h}
}

type CreateInput struct {
	CNPJ        string
	Email       string
	Name        string
	FantasyName *string
	Phone       *string
	Address     *address.CreateInput
}

func (s *Service) Create(ctx context.Context, in CreateInput) (string, error) {
	id := uuid.NewString()
	enterprise := domainenterprise.NewEnterprise(id, in.CNPJ, in.Email, in.Name)

	if in.FantasyName != nil {
		enterprise.SetFantasyName(in.FantasyName)
	}
	if in.Phone != nil {
		enterprise.SetPhone(in.Phone)
	}

	if err := enterprise.Validate(); err != nil {
		return "", apperr.Wrap(err, apperr.CodeInvalid, "invalid enterprise data")
	}

	if in.Address != nil {
		addressID, err := s.addressService.HandleAddress(ctx, in.Address)
		if err != nil {
			return "", err
		}
		enterprise.SetAddressID(&addressID)
	}

	err := s.repo.Create(ctx, enterprise)
	return id, err
}

func (s *Service) GetByID(ctx context.Context, id string) (*domainenterprise.Enterprise, error) {
	if id == "" {
		return nil, apperr.New(apperr.CodeInvalid, "enterprise id is required")
	}
	return s.repo.GetByID(ctx, id)
}

type UpdateInput struct {
	ID          string
	CNPJ        *string
	Email       *string
	Name        *string
	FantasyName *string
	Phone       *string
	Address     *address.CreateInput
	AddressID   *string
}

func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	enterprise, err := s.repo.GetByID(ctx, in.ID)
	if err != nil {
		return err
	}

	if in.CNPJ != nil {
		enterprise.CNPJ = *in.CNPJ
	}
	if in.Email != nil {
		enterprise.Email = *in.Email
	}
	if in.Name != nil {
		enterprise.Name = *in.Name
	}
	if in.FantasyName != nil {
		enterprise.SetFantasyName(in.FantasyName)
	}
	if in.Phone != nil {
		enterprise.SetPhone(in.Phone)
	}

	if err := enterprise.Validate(); err != nil {
		return apperr.Wrap(err, apperr.CodeInvalid, "invalid enterprise data")
	}

	if in.Address != nil {
		addressID, err := s.addressService.HandleAddress(ctx, in.Address)
		if err != nil {
			return err
		}
		enterprise.SetAddressID(&addressID)
	} else if in.AddressID != nil {
		enterprise.SetAddressID(in.AddressID)
	}

	return s.repo.Update(ctx, enterprise)
}
