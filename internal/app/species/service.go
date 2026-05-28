package species

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainspecies "github.com/ESG-Project/suassu-api/internal/domain/species"
	"github.com/google/uuid"
)

type ServiceInterface interface {
	Create(ctx context.Context, in CreateInput) (string, error)
	GetByID(ctx context.Context, id string) (*types.SpeciesWithLegislation, error)
	GetByScientificName(ctx context.Context, scientificName string) (*types.SpeciesWithLegislation, error)
	GetOrCreate(ctx context.Context, in CreateInput) (*types.SpeciesWithLegislation, error)
	List(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error)
}

type Service struct {
	repo Repo
}

func NewService(r Repo) *Service {
	return &Service{repo: r}
}

type CreateInput struct {
	ScientificName      string
	Family              string
	PopularName         *string
	Habit               *string
	LawScope            string
	LawID               *string
	IsLawActive         bool
	SpeciesFormFactor   float64
	IsSpeciesProtected  bool
	SpeciesThreatStatus string
	SpeciesOrigin       string
	SuccessionalEcology string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (string, error) {
	if in.ScientificName == "" || in.Family == "" {
		return "", apperr.New(apperr.CodeInvalid, "missing required fields")
	}

	// Criar espécie primeiro
	speciesID := uuid.NewString()
	species := domainspecies.NewSpecies(
		speciesID,
		in.ScientificName,
		in.Family,
	)

	if in.PopularName != nil {
		species.SetPopularName(in.PopularName)
	}

	if in.Habit != nil {
		species.SetHabit(in.Habit)
	}

	if err := species.Validate(); err != nil {
		return "", apperr.Wrap(err, apperr.CodeInvalid, "invalid species data")
	}

	if err := s.repo.CreateSpecies(ctx, species); err != nil {
		return "", apperr.Wrap(err, apperr.CodeInternal, "failed to create species")
	}

	// Criar legislação associada à espécie
	legislationID := uuid.NewString()
	legislation := domainspecies.NewSpeciesLegislation(
		legislationID,
		in.LawScope,
		in.LawID,
		in.IsLawActive,
		in.SpeciesFormFactor,
		in.IsSpeciesProtected,
		in.SpeciesThreatStatus,
		in.SpeciesOrigin,
		in.SuccessionalEcology,
		&speciesID,
	)

	if err := legislation.Validate(); err != nil {
		return "", apperr.Wrap(err, apperr.CodeInvalid, "invalid legislation data")
	}

	if err := s.repo.CreateLegislation(ctx, legislation); err != nil {
		return "", apperr.Wrap(err, apperr.CodeInternal, "failed to create species legislation")
	}

	return speciesID, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*types.SpeciesWithLegislation, error) {
	species, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "species not found")
	}
	return species, nil
}

func (s *Service) GetByScientificName(ctx context.Context, scientificName string) (*types.SpeciesWithLegislation, error) {
	species, err := s.repo.GetByScientificName(ctx, scientificName)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "species not found")
	}
	return species, nil
}

// GetOrCreate busca uma espécie pelo nome científico ou cria se não existir
func (s *Service) GetOrCreate(ctx context.Context, in CreateInput) (*types.SpeciesWithLegislation, error) {
	// Tentar buscar primeiro
	species, err := s.repo.GetByScientificName(ctx, in.ScientificName)
	if err == nil {
		return species, nil
	}

	// Verificar se o erro é "not found" - se for outro erro, retornar
	if apperr.CodeOf(err) != apperr.CodeNotFound {
		return nil, err
	}

	// Se não encontrou, criar
	id, err := s.Create(ctx, in)
	if err != nil {
		return nil, err
	}

	// Buscar a espécie recém-criada
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error) {
	// Sem limite ou limite muito alto significa retornar todas
	if limit <= 0 {
		limit = 999999
	}

	return s.repo.List(ctx, limit, offset)
}
