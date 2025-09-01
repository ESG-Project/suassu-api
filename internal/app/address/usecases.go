package address

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
	"github.com/google/uuid"
)

type CreateInput struct {
	ZipCode      string
	State        string
	City         string
	Neighborhood string
	Street       string
	Num          string
	Latitude     *string
	Longitude    *string
	AddInfo      *string
}

type Service struct {
	repo   Repo
	hasher Hasher
}

func NewService(repo Repo, hasher Hasher) *Service {
	return &Service{
		repo:   repo,
		hasher: hasher,
	}
}

func (s *Service) HandleAddress(ctx context.Context, in *CreateInput) (string, error) {
	if in.Latitude != nil && *in.Latitude == "" {
		in.Latitude = nil
	}
	if in.Longitude != nil && *in.Longitude == "" {
		in.Longitude = nil
	}
	if in.AddInfo != nil && *in.AddInfo == "" {
		in.AddInfo = nil
	}
	// REGRA DE NEGÓCIO: Verificar se endereço existe antes de criar
	searchParams := domainaddress.NewSearchParams(
		in.ZipCode,
		in.State,
		in.City,
		in.Neighborhood,
		in.Street,
		in.Num,
		in.Latitude,
		in.Longitude,
		in.AddInfo,
	)
	existingAddr, err := s.repo.FindByDetails(ctx, searchParams)

	// Se encontrou endereço existente, retorna o ID
	if err == nil && existingAddr != nil {
		return existingAddr.ID, nil
	}

	// Se não encontrou, cria novo endereço
	newAddrID := uuid.NewString()
	newAddr := domainaddress.NewAddress(newAddrID, in.ZipCode, in.State, in.City, in.Neighborhood, in.Street, in.Num)

	if in.Latitude != nil {
		newAddr.SetLatitude(in.Latitude)
	}
	if in.Longitude != nil {
		newAddr.SetLongitude(in.Longitude)
	}
	if in.AddInfo != nil {
		newAddr.SetAddInfo(in.AddInfo)
	}

	// Validate address before saving
	if err := newAddr.Validate(); err != nil {
		return "", apperr.Wrap(err, apperr.CodeInvalid, "invalid address data")
	}

	err = s.repo.Create(ctx, newAddr)
	if err != nil {
		return "", err
	}

	return newAddrID, nil
}
