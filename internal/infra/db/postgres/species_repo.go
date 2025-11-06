package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainspecies "github.com/ESG-Project/suassu-api/internal/domain/species"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type SpeciesRepo struct {
	q *sqlc.Queries
}

func NewSpeciesRepoFrom(d dbtx) *SpeciesRepo {
	return &SpeciesRepo{q: sqlc.New(d)}
}

func NewSpeciesRepo(db *sql.DB) *SpeciesRepo {
	return &SpeciesRepo{q: sqlc.New(db)}
}

func (r *SpeciesRepo) CreateLegislation(ctx context.Context, sl *domainspecies.SpeciesLegislation) error {
	_, err := r.q.CreateSpeciesLegislation(ctx, sqlc.CreateSpeciesLegislationParams{
		ID:                  sl.ID,
		LawScope:            sqlc.LawScope(sl.LawScope),
		LawID:               sl.LawID,
		IsLawActive:         sl.IsLawActive,
		SpeciesFormFactor:   utils.Float64ToString(sl.SpeciesFormFactor),
		IsSpeciesProtected:  sl.IsSpeciesProtected,
		SpeciesThreatStatus: sqlc.ThreatStatus(sl.SpeciesThreatStatus),
		SuccessionalEcology: sqlc.OriginType(sl.SpeciesOrigin),
		CreatedAt:           sl.CreatedAt,
		UpdatedAt:           sl.UpdatedAt,
	})
	return err
}

func (r *SpeciesRepo) CreateSpecies(ctx context.Context, s *domainspecies.Species) error {
	_, err := r.q.CreateSpecies(ctx, sqlc.CreateSpeciesParams{
		ID:              s.ID,
		ScientificName:  s.ScientificName,
		Family:          s.Family,
		PopularName:     utils.ToNullString(s.PopularName),
		SpeciesDetailID: s.SpeciesDetailID,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	})
	return err
}

func (r *SpeciesRepo) GetByID(ctx context.Context, id string) (*types.SpeciesWithLegislation, error) {
	row, err := r.q.GetSpeciesByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "species not found")
		}
		return nil, err
	}

	formFactor, _ := utils.StringToFloat64(row.SpeciesFormFactor)

	return &types.SpeciesWithLegislation{
		ID:                  row.ID,
		ScientificName:      row.ScientificName,
		Family:              row.Family,
		PopularName:         utils.FromNullString(row.PopularName),
		SpeciesDetailID:     row.SpeciesDetailID,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
		LawScope:            string(row.LawScope),
		LawID:               row.LawID,
		IsLawActive:         row.IsLawActive,
		SpeciesFormFactor:   formFactor,
		IsSpeciesProtected:  row.IsSpeciesProtected,
		SpeciesThreatStatus: string(row.SpeciesThreatStatus),
		SpeciesOrigin:       string(row.SuccessionalEcology),
	}, nil
}

func (r *SpeciesRepo) GetByScientificName(ctx context.Context, scientificName string) (*types.SpeciesWithLegislation, error) {
	row, err := r.q.GetSpeciesByScientificName(ctx, scientificName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "species not found")
		}
		return nil, err
	}

	formFactor, _ := utils.StringToFloat64(row.SpeciesFormFactor)

	return &types.SpeciesWithLegislation{
		ID:                  row.ID,
		ScientificName:      row.ScientificName,
		Family:              row.Family,
		PopularName:         utils.FromNullString(row.PopularName),
		SpeciesDetailID:     row.SpeciesDetailID,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
		LawScope:            string(row.LawScope),
		LawID:               row.LawID,
		IsLawActive:         row.IsLawActive,
		SpeciesFormFactor:   formFactor,
		IsSpeciesProtected:  row.IsSpeciesProtected,
		SpeciesThreatStatus: string(row.SpeciesThreatStatus),
		SpeciesOrigin:       string(row.SuccessionalEcology),
	}, nil
}

func (r *SpeciesRepo) List(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error) {
	rows, err := r.q.ListSpecies(ctx, sqlc.ListSpeciesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*types.SpeciesWithLegislation, 0, len(rows))
	for _, row := range rows {
		formFactor, _ := utils.StringToFloat64(row.SpeciesFormFactor)

		result = append(result, &types.SpeciesWithLegislation{
			ID:                  row.ID,
			ScientificName:      row.ScientificName,
			Family:              row.Family,
			PopularName:         utils.FromNullString(row.PopularName),
			SpeciesDetailID:     row.SpeciesDetailID,
			CreatedAt:           row.CreatedAt,
			UpdatedAt:           row.UpdatedAt,
			LawScope:            string(row.LawScope),
			LawID:               row.LawID,
			IsLawActive:         row.IsLawActive,
			SpeciesFormFactor:   formFactor,
			IsSpeciesProtected:  row.IsSpeciesProtected,
			SpeciesThreatStatus: string(row.SpeciesThreatStatus),
			SpeciesOrigin:       string(row.SuccessionalEcology),
		})
	}

	return result, nil
}

func (r *SpeciesRepo) UpdateSpecies(ctx context.Context, s *domainspecies.Species) error {
	return r.q.UpdateSpecies(ctx, sqlc.UpdateSpeciesParams{
		ID:             s.ID,
		ScientificName: s.ScientificName,
		Family:         s.Family,
		PopularName:    utils.ToNullString(s.PopularName),
		UpdatedAt:      s.UpdatedAt,
	})
}

func (r *SpeciesRepo) UpdateLegislation(ctx context.Context, sl *domainspecies.SpeciesLegislation) error {
	return r.q.UpdateSpeciesLegislation(ctx, sqlc.UpdateSpeciesLegislationParams{
		ID:                  sl.ID,
		LawScope:            sqlc.LawScope(sl.LawScope),
		LawID:               sl.LawID,
		IsLawActive:         sl.IsLawActive,
		SpeciesFormFactor:   utils.Float64ToString(sl.SpeciesFormFactor),
		IsSpeciesProtected:  sl.IsSpeciesProtected,
		SpeciesThreatStatus: sqlc.ThreatStatus(sl.SpeciesThreatStatus),
		SuccessionalEcology: sqlc.OriginType(sl.SpeciesOrigin),
		UpdatedAt:           sl.UpdatedAt,
	})
}
