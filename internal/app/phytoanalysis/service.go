package phytoanalysis

import (
	"context"
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/species"
	"github.com/ESG-Project/suassu-api/internal/app/specimen"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainphyto "github.com/ESG-Project/suassu-api/internal/domain/phytoanalysis"
	postgres "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
	"github.com/google/uuid"
)

type ServiceInterface interface {
	Create(ctx context.Context, in CreateInput) (string, error)
	GetByID(ctx context.Context, id string) (*types.PhytoAnalysisWithProject, error)
	GetWithSpecimens(ctx context.Context, id string) (*types.PhytoAnalysisComplete, error)
	ListByProject(ctx context.Context, projectID string) ([]*types.PhytoAnalysisWithProject, error)
	ListByEnterprise(ctx context.Context, enterpriseID string) ([]*types.PhytoAnalysisWithProject, error)
	ListAll(ctx context.Context, limit, offset int32) ([]*types.PhytoAnalysisWithProject, error)
	Update(ctx context.Context, id string, in UpdateInput) error
	Delete(ctx context.Context, id string) error
}

type Service struct {
	repo Repo
	txm  postgres.TxManagerInterface
}

func NewService(r Repo, txm postgres.TxManagerInterface) *Service {
	return &Service{
		repo: r,
		txm:  txm,
	}
}

type CreateInput struct {
	Title           string
	InitialDate     time.Time
	PortionQuantity int
	PortionArea     float64
	TotalArea       float64
	SampledArea     float64
	Description     *string
	ProjectID       string
	Specimens       []SpecimenInput
}

type SpecimenInput struct {
	Portion      string
	Height       float64
	Cap1         float64
	Cap2         *float64
	Cap3         *float64
	Cap4         *float64
	Cap5         *float64
	Cap6         *float64
	RegisterDate time.Time
	// Dados da espécie - buscar pelo nome científico
	ScientificName string // Nome científico da espécie (obrigatório)
}

type UpdateInput struct {
	Title           string
	InitialDate     time.Time
	PortionQuantity int
	PortionArea     float64
	TotalArea       float64
	SampledArea     float64
	Description     *string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (string, error) {
	if in.Title == "" || in.ProjectID == "" {
		return "", apperr.New(apperr.CodeInvalid, "missing required fields")
	}

	// Criar PhytoAnalysis e Specimens em uma transação
	phytoID := uuid.NewString()

	if s.txm == nil {
		return "", apperr.New(apperr.CodeInvalid, "transaction manager required")
	}

	err := s.txm.RunInTx(ctx, func(repos postgres.Repos) error {
		// 1. Criar PhytoAnalysis
		phyto := domainphyto.NewPhytoAnalysis(
			phytoID,
			in.Title,
			in.InitialDate,
			in.PortionQuantity,
			in.PortionArea,
			in.TotalArea,
			in.SampledArea,
			in.ProjectID,
		)

		if in.Description != nil {
			phyto.SetDescription(in.Description)
		}

		if err := phyto.Validate(); err != nil {
			return apperr.Wrap(err, apperr.CodeInvalid, "invalid phyto analysis data")
		}

		if err := repos.PhytoAnalyses().Create(ctx, phyto); err != nil {
			return err
		}

		// 2. Criar services com repos transacionais
		speciesRepo := repos.Species()
		specimenRepo := repos.Specimens()

		speciesSvc := species.NewService(speciesRepo)
		specimenSvc := specimen.NewService(specimenRepo)

		// 3. Criar specimens
		for _, specimenIn := range in.Specimens {
			// Buscar espécie pelo nome científico
			speciesData, err := speciesSvc.GetByScientificName(ctx, specimenIn.ScientificName)
			if err != nil {
				return apperr.Wrap(err, apperr.CodeNotFound, "species not found with scientific name: "+specimenIn.ScientificName)
			}

			// Criar specimen
			specimenInput := specimen.CreateInput{
				Portion:         specimenIn.Portion,
				Height:          specimenIn.Height,
				Cap1:            specimenIn.Cap1,
				Cap2:            specimenIn.Cap2,
				Cap3:            specimenIn.Cap3,
				Cap4:            specimenIn.Cap4,
				Cap5:            specimenIn.Cap5,
				Cap6:            specimenIn.Cap6,
				RegisterDate:    specimenIn.RegisterDate,
				PhytoAnalysisID: phytoID,
				SpecieID:        speciesData.ID,
			}

			if _, err := specimenSvc.Create(ctx, specimenInput); err != nil {
				return apperr.Wrap(err, apperr.CodeInvalid, "failed to create specimen")
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return phytoID, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*types.PhytoAnalysisWithProject, error) {
	phyto, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "phyto analysis not found")
	}
	return phyto, nil
}

func (s *Service) GetWithSpecimens(ctx context.Context, id string) (*types.PhytoAnalysisComplete, error) {
	phyto, err := s.repo.GetWithSpecimens(ctx, id)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "phyto analysis not found")
	}
	return phyto, nil
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]*types.PhytoAnalysisWithProject, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) ListByEnterprise(ctx context.Context, enterpriseID string) ([]*types.PhytoAnalysisWithProject, error) {
	return s.repo.ListByEnterprise(ctx, enterpriseID)
}

func (s *Service) ListAll(ctx context.Context, limit, offset int32) ([]*types.PhytoAnalysisWithProject, error) {
	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	return s.repo.ListAll(ctx, limit, offset)
}

func (s *Service) Update(ctx context.Context, id string, in UpdateInput) error {
	if in.Title == "" {
		return apperr.New(apperr.CodeInvalid, "missing required fields")
	}

	// Para update, usamos um projectID dummy já que ele não é alterado
	// A validação completa seria feita no repo que não altera o projectID
	phyto := domainphyto.NewPhytoAnalysis(
		id,
		in.Title,
		in.InitialDate,
		in.PortionQuantity,
		in.PortionArea,
		in.TotalArea,
		in.SampledArea,
		"dummy-project-id", // projectID não é alterado no update
	)

	if in.Description != nil {
		phyto.SetDescription(in.Description)
	}

	// Validar apenas os campos que podem ser atualizados
	if in.PortionQuantity <= 0 {
		return apperr.New(apperr.CodeInvalid, "portion quantity must be positive")
	}
	if in.PortionArea <= 0 {
		return apperr.New(apperr.CodeInvalid, "portion area must be positive")
	}
	if in.TotalArea <= 0 {
		return apperr.New(apperr.CodeInvalid, "total area must be positive")
	}
	if in.SampledArea <= 0 {
		return apperr.New(apperr.CodeInvalid, "sampled area must be positive")
	}

	return s.repo.Update(ctx, phyto)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
