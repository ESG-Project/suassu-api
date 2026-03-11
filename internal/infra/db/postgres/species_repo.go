package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainspecies "github.com/ESG-Project/suassu-api/internal/domain/species"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type SpeciesRepo struct {
	q  *sqlc.Queries
	db dbtx
}

func NewSpeciesRepoFrom(d dbtx) *SpeciesRepo {
	return &SpeciesRepo{q: sqlc.New(d), db: d}
}

func NewSpeciesRepo(db *sql.DB) *SpeciesRepo {
	return &SpeciesRepo{q: sqlc.New(db), db: db}
}

func (r *SpeciesRepo) CreateSpecies(ctx context.Context, s *domainspecies.Species) error {
	_, err := r.q.CreateSpecies(ctx, sqlc.CreateSpeciesParams{
		ID:             s.ID,
		ScientificName: s.ScientificName,
		Family:         s.Family,
		PopularName:    utils.ToNullString(s.PopularName),
		Habit:          utils.ToNullSpeciesHabit(s.Habit),
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	})
	return err
}

func (r *SpeciesRepo) CreateLegislation(ctx context.Context, sl *domainspecies.SpeciesLegislation) error {
	_, err := r.q.CreateSpeciesLegislation(ctx, sqlc.CreateSpeciesLegislationParams{
		ID:                  sl.ID,
		LawScope:            sqlc.LawScope(sl.LawScope),
		LawID:               utils.ToNullString(sl.LawID),
		IsLawActive:         sl.IsLawActive,
		SpeciesFormFactor:   utils.Float64ToString(sl.SpeciesFormFactor),
		IsSpeciesProtected:  sl.IsSpeciesProtected,
		SpeciesThreatStatus: sqlc.ThreatStatus(sl.SpeciesThreatStatus),
		SpeciesOrigin:       sqlc.OriginType(sl.SpeciesOrigin),
		SuccessionalEcology: sqlc.SpeciesSuccessionalEcology(sl.SuccessionalEcology),
		SpeciesID:           utils.ToNullString(sl.SpeciesID),
		CreatedAt:           sl.CreatedAt,
		UpdatedAt:           sl.UpdatedAt,
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

	// Buscar legislações associadas
	legislations, err := r.q.GetSpeciesLegislationsBySpeciesID(ctx, utils.StringToNullString(id))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Converter legislações para o tipo apropriado
	legislationData := make([]types.LegislationData, 0, len(legislations))
	for _, leg := range legislations {
		formFactor, _ := utils.StringToFloat64(leg.SpeciesFormFactor)
		legislationData = append(legislationData, types.LegislationData{
			ID:                  leg.ID,
			LawScope:            string(leg.LawScope),
			LawID:               utils.FromNullString(leg.LawID),
			IsLawActive:         leg.IsLawActive,
			SpeciesFormFactor:   formFactor,
			IsSpeciesProtected:  leg.IsSpeciesProtected,
			SpeciesThreatStatus: string(leg.SpeciesThreatStatus),
			SpeciesOrigin:       string(leg.SpeciesOrigin),
			SuccessionalEcology: string(leg.SuccessionalEcology),
			SpeciesID:           utils.FromNullString(leg.SpeciesID),
			CreatedAt:           leg.CreatedAt,
			UpdatedAt:           leg.UpdatedAt,
		})
	}

	return &types.SpeciesWithLegislation{
		ID:             row.ID,
		ScientificName: row.ScientificName,
		Family:         row.Family,
		PopularName:    utils.FromNullString(row.PopularName),
		Habit:          utils.FromNullSpeciesHabit(row.Habit),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		Legislations:   legislationData,
	}, nil
}

// GetMapByScientificNames busca várias espécies de uma vez e retorna
// um mapa scientificName -> speciesID. Uma única query SQL.
func (r *SpeciesRepo) GetMapByScientificNames(ctx context.Context, names []string) (map[string]string, error) {
	if len(names) == 0 {
		return make(map[string]string), nil
	}

	unique := make(map[string]struct{}, len(names))
	for _, n := range names {
		unique[strings.TrimSpace(n)] = struct{}{}
	}

	placeholders := make([]string, 0, len(unique))
	args := make([]interface{}, 0, len(unique))
	i := 1
	for name := range unique {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, name)
		i++
	}

	query := fmt.Sprintf(
		"SELECT id, scientific_name FROM public.species WHERE scientific_name IN (%s)",
		strings.Join(placeholders, ", "),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string, len(unique))
	for rows.Next() {
		var id, scientificName string
		if err := rows.Scan(&id, &scientificName); err != nil {
			return nil, err
		}
		result[scientificName] = id
	}

	return result, rows.Err()
}

func (r *SpeciesRepo) GetByScientificName(ctx context.Context, scientificName string) (*types.SpeciesWithLegislation, error) {
	row, err := r.q.GetSpeciesByScientificName(ctx, scientificName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "species not found")
		}
		return nil, err
	}

	// Buscar legislações associadas
	legislations, err := r.q.GetSpeciesLegislationsBySpeciesID(ctx, utils.StringToNullString(row.ID))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Converter legislações para o tipo apropriado
	legislationData := make([]types.LegislationData, 0, len(legislations))
	for _, leg := range legislations {
		formFactor, _ := utils.StringToFloat64(leg.SpeciesFormFactor)
		legislationData = append(legislationData, types.LegislationData{
			ID:                  leg.ID,
			LawScope:            string(leg.LawScope),
			LawID:               utils.FromNullString(leg.LawID),
			IsLawActive:         leg.IsLawActive,
			SpeciesFormFactor:   formFactor,
			IsSpeciesProtected:  leg.IsSpeciesProtected,
			SpeciesThreatStatus: string(leg.SpeciesThreatStatus),
			SpeciesOrigin:       string(leg.SpeciesOrigin),
			SuccessionalEcology: string(leg.SuccessionalEcology),
			SpeciesID:           utils.FromNullString(leg.SpeciesID),
			CreatedAt:           leg.CreatedAt,
			UpdatedAt:           leg.UpdatedAt,
		})
	}

	return &types.SpeciesWithLegislation{
		ID:             row.ID,
		ScientificName: row.ScientificName,
		Family:         row.Family,
		PopularName:    utils.FromNullString(row.PopularName),
		Habit:          utils.FromNullSpeciesHabit(row.Habit),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		Legislations:   legislationData,
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
		// Buscar legislações associadas a cada espécie
		legislations, err := r.q.GetSpeciesLegislationsBySpeciesID(ctx, utils.StringToNullString(row.ID))
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		// Converter legislações para o tipo apropriado
		legislationData := make([]types.LegislationData, 0, len(legislations))
		for _, leg := range legislations {
			formFactor, _ := utils.StringToFloat64(leg.SpeciesFormFactor)
			legislationData = append(legislationData, types.LegislationData{
				ID:                  leg.ID,
				LawScope:            string(leg.LawScope),
				LawID:               utils.FromNullString(leg.LawID),
				IsLawActive:         leg.IsLawActive,
				SpeciesFormFactor:   formFactor,
				IsSpeciesProtected:  leg.IsSpeciesProtected,
				SpeciesThreatStatus: string(leg.SpeciesThreatStatus),
				SpeciesOrigin:       string(leg.SpeciesOrigin),
				SuccessionalEcology: string(leg.SuccessionalEcology),
				SpeciesID:           utils.FromNullString(leg.SpeciesID),
				CreatedAt:           leg.CreatedAt,
				UpdatedAt:           leg.UpdatedAt,
			})
		}

		result = append(result, &types.SpeciesWithLegislation{
			ID:             row.ID,
			ScientificName: row.ScientificName,
			Family:         row.Family,
			PopularName:    utils.FromNullString(row.PopularName),
			Habit:          utils.FromNullSpeciesHabit(row.Habit),
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
			Legislations:   legislationData,
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
		Habit:          utils.ToNullSpeciesHabit(s.Habit),
		UpdatedAt:      s.UpdatedAt,
	})
}

func (r *SpeciesRepo) UpdateLegislation(ctx context.Context, sl *domainspecies.SpeciesLegislation) error {
	return r.q.UpdateSpeciesLegislation(ctx, sqlc.UpdateSpeciesLegislationParams{
		ID:                  sl.ID,
		LawScope:            sqlc.LawScope(sl.LawScope),
		LawID:               utils.ToNullString(sl.LawID),
		IsLawActive:         sl.IsLawActive,
		SpeciesFormFactor:   utils.Float64ToString(sl.SpeciesFormFactor),
		IsSpeciesProtected:  sl.IsSpeciesProtected,
		SpeciesThreatStatus: sqlc.ThreatStatus(sl.SpeciesThreatStatus),
		SpeciesOrigin:       sqlc.OriginType(sl.SpeciesOrigin),
		SuccessionalEcology: sqlc.SpeciesSuccessionalEcology(sl.SuccessionalEcology),
		UpdatedAt:           sl.UpdatedAt,
	})
}
