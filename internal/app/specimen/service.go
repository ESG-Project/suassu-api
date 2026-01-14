package specimen

import (
	"context"
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainspecimen "github.com/ESG-Project/suassu-api/internal/domain/specimen"
	"github.com/google/uuid"
)

type ServiceInterface interface {
	Create(ctx context.Context, in CreateInput) (string, error)
	GetByID(ctx context.Context, id string) (*types.SpecimenWithSpecies, error)
	ListByPhytoAnalysis(ctx context.Context, phytoAnalysisID string) ([]*types.SpecimenWithSpecies, error)
	Update(ctx context.Context, id string, in UpdateInput) error
	Delete(ctx context.Context, id string) error
}

type Service struct {
	repo Repo
}

func NewService(r Repo) *Service {
	return &Service{repo: r}
}

type CreateInput struct {
	Portion         string
	Height          float64
	Cap1            float64
	Cap2            *float64
	Cap3            *float64
	Cap4            *float64
	Cap5            *float64
	Cap6            *float64
	RegisterDate    time.Time
	PhytoAnalysisID string
	SpecieID        string
}

type UpdateInput struct {
	Portion      string
	Height       float64
	Cap1         float64
	Cap2         *float64
	Cap3         *float64
	Cap4         *float64
	Cap5         *float64
	Cap6         *float64
	RegisterDate time.Time
	SpecieID     string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (string, error) {
	if in.Portion == "" || in.PhytoAnalysisID == "" || in.SpecieID == "" {
		return "", apperr.New(apperr.CodeInvalid, "missing required fields")
	}

	id := uuid.NewString()
	specimen := domainspecimen.NewSpecimen(
		id,
		in.Portion,
		in.Height,
		in.Cap1,
		in.RegisterDate,
		in.PhytoAnalysisID,
		in.SpecieID,
	)

	specimen.SetOptionalCaps(in.Cap2, in.Cap3, in.Cap4, in.Cap5, in.Cap6)

	if err := specimen.Validate(); err != nil {
		return "", apperr.Wrap(err, apperr.CodeInvalid, "invalid specimen data")
	}

	if err := s.repo.Create(ctx, specimen); err != nil {
		return "", err
	}

	return id, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*types.SpecimenWithSpecies, error) {
	specimen, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "specimen not found")
	}
	return specimen, nil
}

func (s *Service) ListByPhytoAnalysis(ctx context.Context, phytoAnalysisID string) ([]*types.SpecimenWithSpecies, error) {
	return s.repo.ListByPhytoAnalysis(ctx, phytoAnalysisID)
}

func (s *Service) Update(ctx context.Context, id string, in UpdateInput) error {
	if in.Portion == "" || in.SpecieID == "" {
		return apperr.New(apperr.CodeInvalid, "missing required fields")
	}

	// Buscar o specimen existente para pegar o phytoAnalysisID e CreatedAt (não podem ser alterados)
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return apperr.Wrap(err, apperr.CodeNotFound, "specimen not found")
	}

	specimen := domainspecimen.NewSpecimen(
		id,
		in.Portion,
		in.Height,
		in.Cap1,
		in.RegisterDate,
		existing.PhytoAnalysisID,
		in.SpecieID,
	)

	// Manter o CreatedAt original e atualizar apenas o UpdatedAt
	specimen.CreatedAt = existing.CreatedAt
	specimen.UpdatedAt = time.Now()

	specimen.SetOptionalCaps(in.Cap2, in.Cap3, in.Cap4, in.Cap5, in.Cap6)

	if err := specimen.Validate(); err != nil {
		return apperr.Wrap(err, apperr.CodeInvalid, "invalid specimen data")
	}

	return s.repo.Update(ctx, specimen)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
