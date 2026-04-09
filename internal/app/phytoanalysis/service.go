package phytoanalysis

import (
	"context"
	"strings"
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainphyto "github.com/ESG-Project/suassu-api/internal/domain/phytoanalysis"
	domainspecimen "github.com/ESG-Project/suassu-api/internal/domain/specimen"
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
	Description     *string
}

type specimenRow struct {
	RowNumber int
	Specimen  SpecimenInput
}

type invalidSpecimenRow struct {
	RowNumber int      `json:"rowNumber"`
	Errors    []string `json:"errors"`
}

func isBlankSpecimenInput(sp SpecimenInput) bool {
	return strings.TrimSpace(sp.Portion) == "" &&
		strings.TrimSpace(sp.ScientificName) == "" &&
		sp.Height == 0 &&
		sp.Cap1 == 0 &&
		sp.RegisterDate.IsZero() &&
		sp.Cap2 == nil &&
		sp.Cap3 == nil &&
		sp.Cap4 == nil &&
		sp.Cap5 == nil &&
		sp.Cap6 == nil
}

func normalizeAndValidateSpecimens(specimens []SpecimenInput) ([]specimenRow, []invalidSpecimenRow) {
	rows := make([]specimenRow, 0, len(specimens))
	invalidRows := make([]invalidSpecimenRow, 0)

	for i, sp := range specimens {
		rowNumber := i + 1
		if isBlankSpecimenInput(sp) {
			continue
		}

		normalized := sp
		normalized.Portion = strings.TrimSpace(sp.Portion)
		normalized.ScientificName = strings.TrimSpace(sp.ScientificName)

		errorsByRow := make([]string, 0, 5)
		if normalized.Portion == "" {
			errorsByRow = append(errorsByRow, "portion is required")
		}
		if normalized.Height <= 0 {
			errorsByRow = append(errorsByRow, "height must be positive")
		}
		if normalized.Cap1 <= 0 {
			errorsByRow = append(errorsByRow, "cap1 must be positive")
		}
		if normalized.RegisterDate.IsZero() {
			errorsByRow = append(errorsByRow, "register date is required")
		}
		if normalized.ScientificName == "" {
			errorsByRow = append(errorsByRow, "scientific name is required")
		}

		if len(errorsByRow) > 0 {
			invalidRows = append(invalidRows, invalidSpecimenRow{
				RowNumber: rowNumber,
				Errors:    errorsByRow,
			})
			continue
		}

		rows = append(rows, specimenRow{RowNumber: rowNumber, Specimen: normalized})
	}

	return rows, invalidRows
}

func calcSampledAreaHa(portionArea float64, portionQuantity int) float64 {
	if portionArea <= 0 || portionQuantity <= 0 {
		return 0
	}
	plotsAreaM2 := portionArea * float64(portionQuantity)
	return plotsAreaM2 / 10000.0
}

func (s *Service) Create(ctx context.Context, in CreateInput) (string, error) {
	if in.Title == "" || in.ProjectID == "" {
		return "", apperr.New(apperr.CodeInvalid, "missing required fields")
	}

	phytoID := uuid.NewString()

	if s.txm == nil {
		return "", apperr.New(apperr.CodeInvalid, "transaction manager required")
	}

	err := s.txm.RunInTx(ctx, func(repos postgres.Repos) error {
		sampledAreaHa := calcSampledAreaHa(in.PortionArea, in.PortionQuantity)
		if sampledAreaHa <= 0 {
			return apperr.New(apperr.CodeInvalid, "sampled area must be positive")
		}

		rows, invalidRows := normalizeAndValidateSpecimens(in.Specimens)
		if len(invalidRows) > 0 {
			return apperr.WithFields(
				apperr.New(apperr.CodeInvalid, "invalid specimen rows"),
				map[string]any{"invalidRows": invalidRows},
			)
		}

		phyto := domainphyto.NewPhytoAnalysis(
			phytoID,
			in.Title,
			in.InitialDate,
			in.PortionQuantity,
			in.PortionArea,
			in.TotalArea,
			sampledAreaHa,
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

		if len(rows) == 0 {
			return nil
		}

		// Batch: coletar nomes científicos únicos (trim) e buscar todos de uma vez
		uniqueNames := make([]string, 0, len(rows))
		seen := make(map[string]bool, len(rows))
		for _, row := range rows {
			name := row.Specimen.ScientificName
			if !seen[name] {
				seen[name] = true
				uniqueNames = append(uniqueNames, name)
			}
		}

		speciesMap, err := repos.Species().GetMapByScientificNames(ctx, uniqueNames)
		if err != nil {
			return apperr.Wrap(err, apperr.CodeInternal, "failed to fetch species")
		}

		missingSpeciesRows := make([]invalidSpecimenRow, 0)
		for _, row := range rows {
			name := row.Specimen.ScientificName
			if _, ok := speciesMap[name]; !ok {
				missingSpeciesRows = append(missingSpeciesRows, invalidSpecimenRow{
					RowNumber: row.RowNumber,
					Errors:    []string{"species not found with scientific name: " + name},
				})
			}
		}

		if len(missingSpeciesRows) > 0 {
			return apperr.WithFields(
				apperr.New(apperr.CodeInvalid, "invalid specimen rows"),
				map[string]any{"invalidRows": missingSpeciesRows},
			)
		}

		// Batch: construir todas as entidades de specimen e inserir de uma vez
		domainSpecimens := make([]*domainspecimen.Specimen, 0, len(rows))
		for _, row := range rows {
			sp := row.Specimen
			specieID := speciesMap[sp.ScientificName]

			s := domainspecimen.NewSpecimen(
				uuid.NewString(),
				sp.Portion,
				sp.Height,
				sp.Cap1,
				sp.RegisterDate,
				phytoID,
				specieID,
			)
			s.SetOptionalCaps(sp.Cap2, sp.Cap3, sp.Cap4, sp.Cap5, sp.Cap6)

			if err := s.Validate(); err != nil {
				return apperr.WithFields(
					apperr.New(apperr.CodeInvalid, "invalid specimen rows"),
					map[string]any{
						"invalidRows": []invalidSpecimenRow{{
							RowNumber: row.RowNumber,
							Errors:    []string{err.Error()},
						}},
					},
				)
			}

			domainSpecimens = append(domainSpecimens, s)
		}

		if err := repos.Specimens().CreateBatch(ctx, domainSpecimens); err != nil {
			return apperr.Wrap(err, apperr.CodeInvalid, "failed to create specimens")
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
	if strings.TrimSpace(in.Title) == "" {
		return apperr.New(apperr.CodeInvalid, "missing required fields")
	}
	if in.InitialDate.IsZero() {
		return apperr.New(apperr.CodeInvalid, "missing required fields")
	}
	if in.PortionQuantity <= 0 {
		return apperr.New(apperr.CodeInvalid, "portion quantity must be positive")
	}
	if in.PortionArea <= 0 {
		return apperr.New(apperr.CodeInvalid, "portion area must be positive")
	}
	if in.TotalArea <= 0 {
		return apperr.New(apperr.CodeInvalid, "total area must be positive")
	}

	sampledAreaHa := calcSampledAreaHa(in.PortionArea, in.PortionQuantity)
	if sampledAreaHa <= 0 {
		return apperr.New(apperr.CodeInvalid, "sampled area must be positive")
	}

	phyto := &domainphyto.PhytoAnalysis{
		ID:              id,
		Title:           in.Title,
		InitialDate:     in.InitialDate,
		PortionQuantity: in.PortionQuantity,
		PortionArea:     in.PortionArea,
		TotalArea:       in.TotalArea,
		SampledArea:     sampledAreaHa,
		Description:     in.Description,
		UpdatedAt:       time.Now(),
	}

	return s.repo.Update(ctx, phyto)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s.txm == nil {
		return s.repo.Delete(ctx, id)
	}

	return s.txm.RunInTx(ctx, func(repos postgres.Repos) error {
		if err := repos.Specimens().DeleteByPhytoAnalysis(ctx, id); err != nil {
			return err
		}

		return repos.PhytoAnalyses().Delete(ctx, id)
	})
}
