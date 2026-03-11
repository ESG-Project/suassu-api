package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainspecimen "github.com/ESG-Project/suassu-api/internal/domain/specimen"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type SpecimenRepo struct {
	q  *sqlc.Queries
	db dbtx
}

func NewSpecimenRepoFrom(d dbtx) *SpecimenRepo {
	return &SpecimenRepo{q: sqlc.New(d), db: d}
}

func NewSpecimenRepo(db *sql.DB) *SpecimenRepo {
	return &SpecimenRepo{q: sqlc.New(db), db: db}
}

func (r *SpecimenRepo) Create(ctx context.Context, s *domainspecimen.Specimen) error {
	_, err := r.q.CreateSpecimen(ctx, sqlc.CreateSpecimenParams{
		ID:              s.ID,
		Portion:         s.Portion,
		Height:          utils.Float64ToString(s.Height),
		Cap1:            utils.Float64ToString(s.Cap1),
		Cap2:            utils.Float64PtrToString(s.Cap2),
		Cap3:            utils.Float64PtrToString(s.Cap3),
		Cap4:            utils.Float64PtrToString(s.Cap4),
		Cap5:            utils.Float64PtrToString(s.Cap5),
		Cap6:            utils.Float64PtrToString(s.Cap6),
		RegisterDate:    s.RegisterDate,
		PhytoAnalysisID: s.PhytoAnalysisID,
		SpecieID:        s.SpecieID,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	})
	return err
}

func (r *SpecimenRepo) GetByID(ctx context.Context, id string) (*types.SpecimenWithSpecies, error) {
	row, err := r.q.GetSpecimenByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "specimen not found")
		}
		return nil, err
	}

	height, _ := utils.StringToFloat64(row.Height)
	cap1, _ := utils.StringToFloat64(row.Cap1)

	return &types.SpecimenWithSpecies{
		ID:              row.ID,
		Portion:         row.Portion,
		Height:          height,
		Cap1:            cap1,
		Cap2:            utils.NullStringToNullFloat64(row.Cap2),
		Cap3:            utils.NullStringToNullFloat64(row.Cap3),
		Cap4:            utils.NullStringToNullFloat64(row.Cap4),
		Cap5:            utils.NullStringToNullFloat64(row.Cap5),
		Cap6:            utils.NullStringToNullFloat64(row.Cap6),
		RegisterDate:    row.RegisterDate,
		PhytoAnalysisID: row.PhytoAnalysisID,
		SpecieID:        row.SpecieID,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		ScientificName:  row.ScientificName,
		Family:          row.Family,
		PopularName:     utils.FromNullString(row.PopularName),
	}, nil
}

func (r *SpecimenRepo) ListByPhytoAnalysis(ctx context.Context, phytoAnalysisID string) ([]*types.SpecimenWithSpecies, error) {
	rows, err := r.q.ListSpecimensByPhytoAnalysis(ctx, phytoAnalysisID)
	if err != nil {
		return nil, err
	}

	result := make([]*types.SpecimenWithSpecies, 0, len(rows))
	for _, row := range rows {
		height, _ := utils.StringToFloat64(row.Height)
		cap1, _ := utils.StringToFloat64(row.Cap1)

		result = append(result, &types.SpecimenWithSpecies{
			ID:              row.ID,
			Portion:         row.Portion,
			Height:          height,
			Cap1:            cap1,
			Cap2:            utils.NullStringToNullFloat64(row.Cap2),
			Cap3:            utils.NullStringToNullFloat64(row.Cap3),
			Cap4:            utils.NullStringToNullFloat64(row.Cap4),
			Cap5:            utils.NullStringToNullFloat64(row.Cap5),
			Cap6:            utils.NullStringToNullFloat64(row.Cap6),
			RegisterDate:    row.RegisterDate,
			PhytoAnalysisID: row.PhytoAnalysisID,
			SpecieID:        row.SpecieID,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ScientificName:  row.ScientificName,
			Family:          row.Family,
			PopularName:     utils.FromNullString(row.PopularName),
		})
	}

	return result, nil
}

func (r *SpecimenRepo) Update(ctx context.Context, s *domainspecimen.Specimen) error {
	return r.q.UpdateSpecimen(ctx, sqlc.UpdateSpecimenParams{
		ID:           s.ID,
		Portion:      s.Portion,
		Height:       utils.Float64ToString(s.Height),
		Cap1:         utils.Float64ToString(s.Cap1),
		Cap2:         utils.Float64PtrToString(s.Cap2),
		Cap3:         utils.Float64PtrToString(s.Cap3),
		Cap4:         utils.Float64PtrToString(s.Cap4),
		Cap5:         utils.Float64PtrToString(s.Cap5),
		Cap6:         utils.Float64PtrToString(s.Cap6),
		RegisterDate: s.RegisterDate,
		SpecieID:     s.SpecieID,
		UpdatedAt:    s.UpdatedAt,
	})
}

func (r *SpecimenRepo) Delete(ctx context.Context, id string) error {
	return r.q.DeleteSpecimen(ctx, id)
}

func (r *SpecimenRepo) CountByPhytoAnalysis(ctx context.Context, phytoAnalysisID string) (int64, error) {
	count, err := r.q.CountSpecimensByPhytoAnalysis(ctx, phytoAnalysisID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CreateBatch insere múltiplos specimens em uma única query multi-row INSERT.
func (r *SpecimenRepo) CreateBatch(ctx context.Context, specimens []*domainspecimen.Specimen) error {
	if len(specimens) == 0 {
		return nil
	}

	const colsPerRow = 14
	valueGroups := make([]string, 0, len(specimens))
	args := make([]interface{}, 0, len(specimens)*colsPerRow)

	for i, s := range specimens {
		base := i * colsPerRow
		valueGroups = append(valueGroups, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7,
			base+8, base+9, base+10, base+11, base+12, base+13, base+14,
		))
		args = append(args,
			s.ID,
			s.Portion,
			utils.Float64ToString(s.Height),
			utils.Float64ToString(s.Cap1),
			utils.Float64PtrToString(s.Cap2),
			utils.Float64PtrToString(s.Cap3),
			utils.Float64PtrToString(s.Cap4),
			utils.Float64PtrToString(s.Cap5),
			utils.Float64PtrToString(s.Cap6),
			s.RegisterDate,
			s.PhytoAnalysisID,
			s.SpecieID,
			s.CreatedAt,
			s.UpdatedAt,
		)
	}

	query := `INSERT INTO public.specimen (
		id, portion, height, cap1, cap2, cap3, cap4, cap5, cap6,
		register_date, phyto_analysis_id, specie_id, created_at, updated_at
	) VALUES ` + strings.Join(valueGroups, ", ")

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}
