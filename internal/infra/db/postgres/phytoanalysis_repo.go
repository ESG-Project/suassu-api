package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainphyto "github.com/ESG-Project/suassu-api/internal/domain/phytoanalysis"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type PhytoAnalysisRepo struct {
	q *sqlc.Queries
}

func NewPhytoAnalysisRepoFrom(d dbtx) *PhytoAnalysisRepo {
	return &PhytoAnalysisRepo{q: sqlc.New(d)}
}

func NewPhytoAnalysisRepo(db *sql.DB) *PhytoAnalysisRepo {
	return &PhytoAnalysisRepo{q: sqlc.New(db)}
}

func (r *PhytoAnalysisRepo) Create(ctx context.Context, p *domainphyto.PhytoAnalysis) error {
	_, err := r.q.CreatePhytoAnalysis(ctx, sqlc.CreatePhytoAnalysisParams{
		ID:              p.ID,
		Title:           p.Title,
		InitialDate:     p.InitialDate,
		PortionQuantity: int32(p.PortionQuantity),
		PortionArea:     utils.Float64ToString(p.PortionArea),
		TotalArea:       utils.Float64ToString(p.TotalArea),
		SampledArea:     utils.Float64ToString(p.SampledArea),
		Description:     utils.ToNullString(p.Description),
		ProjectID:       p.ProjectID,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	})
	return err
}

func (r *PhytoAnalysisRepo) GetByID(ctx context.Context, id string) (*types.PhytoAnalysisWithProject, error) {
	row, err := r.q.GetPhytoAnalysisByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "phyto analysis not found")
		}
		return nil, err
	}

	portionArea, _ := utils.StringToFloat64(row.PortionArea)
	totalArea, _ := utils.StringToFloat64(row.TotalArea)
	sampledArea, _ := utils.StringToFloat64(row.SampledArea)

	return &types.PhytoAnalysisWithProject{
		ID:              row.ID,
		Title:           row.Title,
		InitialDate:     row.InitialDate,
		PortionQuantity: int(row.PortionQuantity),
		PortionArea:     portionArea,
		TotalArea:       totalArea,
		SampledArea:     sampledArea,
		Description:     utils.FromNullString(row.Description),
		ProjectID:       row.ProjectID,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		ProjectTitle:    row.ProjectTitle,
		ProjectCNPJ:     utils.FromNullString(row.ProjectCnpj),
		ProjectActivity: row.ProjectActivity,
		ProjectClientID: row.ProjectClientID,
	}, nil
}

func (r *PhytoAnalysisRepo) ListByProject(ctx context.Context, projectID string) ([]*types.PhytoAnalysisWithProject, error) {
	rows, err := r.q.ListPhytoAnalysesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	result := make([]*types.PhytoAnalysisWithProject, 0, len(rows))
	for _, row := range rows {
		portionArea, _ := utils.StringToFloat64(row.PortionArea)
		totalArea, _ := utils.StringToFloat64(row.TotalArea)
		sampledArea, _ := utils.StringToFloat64(row.SampledArea)

		result = append(result, &types.PhytoAnalysisWithProject{
			ID:              row.ID,
			Title:           row.Title,
			InitialDate:     row.InitialDate,
			PortionQuantity: int(row.PortionQuantity),
			PortionArea:     portionArea,
			TotalArea:       totalArea,
			SampledArea:     sampledArea,
			Description:     utils.FromNullString(row.Description),
			ProjectID:       row.ProjectID,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ProjectTitle:    row.ProjectTitle,
			ProjectCNPJ:     utils.FromNullString(row.ProjectCnpj),
			ProjectActivity: row.ProjectActivity,
		})
	}

	return result, nil
}

func (r *PhytoAnalysisRepo) ListByEnterprise(ctx context.Context, enterpriseID string) ([]*types.PhytoAnalysisWithProject, error) {
	rows, err := r.q.ListPhytoAnalysesByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	result := make([]*types.PhytoAnalysisWithProject, 0, len(rows))
	for _, row := range rows {
		portionArea, _ := utils.StringToFloat64(row.PortionArea)
		totalArea, _ := utils.StringToFloat64(row.TotalArea)
		sampledArea, _ := utils.StringToFloat64(row.SampledArea)

		result = append(result, &types.PhytoAnalysisWithProject{
			ID:              row.ID,
			Title:           row.Title,
			InitialDate:     row.InitialDate,
			PortionQuantity: int(row.PortionQuantity),
			PortionArea:     portionArea,
			TotalArea:       totalArea,
			SampledArea:     sampledArea,
			Description:     utils.FromNullString(row.Description),
			ProjectID:       row.ProjectID,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ProjectTitle:    row.ProjectTitle,
			ProjectCNPJ:     utils.FromNullString(row.ProjectCnpj),
			ProjectActivity: row.ProjectActivity,
		})
	}

	return result, nil
}

func (r *PhytoAnalysisRepo) ListAll(ctx context.Context, limit, offset int32) ([]*types.PhytoAnalysisWithProject, error) {
	rows, err := r.q.ListAllPhytoAnalyses(ctx, sqlc.ListAllPhytoAnalysesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*types.PhytoAnalysisWithProject, 0, len(rows))
	for _, row := range rows {
		portionArea, _ := utils.StringToFloat64(row.PortionArea)
		totalArea, _ := utils.StringToFloat64(row.TotalArea)
		sampledArea, _ := utils.StringToFloat64(row.SampledArea)

		result = append(result, &types.PhytoAnalysisWithProject{
			ID:              row.ID,
			Title:           row.Title,
			InitialDate:     row.InitialDate,
			PortionQuantity: int(row.PortionQuantity),
			PortionArea:     portionArea,
			TotalArea:       totalArea,
			SampledArea:     sampledArea,
			Description:     utils.FromNullString(row.Description),
			ProjectID:       row.ProjectID,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ProjectTitle:    row.ProjectTitle,
			ProjectCNPJ:     utils.FromNullString(row.ProjectCnpj),
			ProjectActivity: row.ProjectActivity,
			ProjectClientID: row.ProjectClientID,
		})
	}

	return result, nil
}

func (r *PhytoAnalysisRepo) Update(ctx context.Context, p *domainphyto.PhytoAnalysis) error {
	return r.q.UpdatePhytoAnalysis(ctx, sqlc.UpdatePhytoAnalysisParams{
		ID:              p.ID,
		Title:           p.Title,
		InitialDate:     p.InitialDate,
		PortionQuantity: int32(p.PortionQuantity),
		PortionArea:     utils.Float64ToString(p.PortionArea),
		TotalArea:       utils.Float64ToString(p.TotalArea),
		SampledArea:     utils.Float64ToString(p.SampledArea),
		Description:     utils.ToNullString(p.Description),
		UpdatedAt:       p.UpdatedAt,
	})
}

func (r *PhytoAnalysisRepo) Delete(ctx context.Context, id string) error {
	return r.q.DeletePhytoAnalysis(ctx, id)
}

func (r *PhytoAnalysisRepo) GetWithSpecimens(ctx context.Context, id string) (*types.PhytoAnalysisComplete, error) {
	rows, err := r.q.GetPhytoAnalysisWithSpecimens(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "phyto analysis not found")
		}
		return nil, err
	}

	if len(rows) == 0 {
		return nil, apperr.New(apperr.CodeNotFound, "phyto analysis not found")
	}

	// A primeira linha contém os dados do PhytoAnalysis e Project
	firstRow := rows[0]

	portionArea, _ := utils.StringToFloat64(firstRow.PortionArea)
	totalArea, _ := utils.StringToFloat64(firstRow.TotalArea)
	sampledArea, _ := utils.StringToFloat64(firstRow.SampledArea)

	result := &types.PhytoAnalysisComplete{
		ID:              firstRow.PhytoID,
		Title:           firstRow.PhytoTitle,
		InitialDate:     firstRow.InitialDate,
		PortionQuantity: int(firstRow.PortionQuantity),
		PortionArea:     portionArea,
		TotalArea:       totalArea,
		SampledArea:     sampledArea,
		Description:     utils.FromNullString(firstRow.PhytoDescription),
		ProjectID:       firstRow.ProjectID,
		CreatedAt:       firstRow.PhytoCreatedAt,
		UpdatedAt:       firstRow.PhytoUpdatedAt,
		ProjectTitle:    firstRow.ProjectTitle,
		ProjectCNPJ:     utils.FromNullString(firstRow.ProjectCnpj),
		ProjectActivity: firstRow.ProjectActivity,
		ProjectClientID: firstRow.ProjectClientID,
		Specimens:       make([]*types.SpecimenWithSpecies, 0),
	}

	// Agregar specimens
	for _, row := range rows {
		// Se não há specimen (LEFT JOIN vazio), pular
		if !row.SpecimenID.Valid {
			continue
		}

		height := utils.NullStringToNullFloat64(row.Height)
		cap1 := utils.NullStringToNullFloat64(row.Cap1)
		averageDap := utils.NullStringToNullFloat64(row.AverageDap)
		basalArea := utils.NullStringToNullFloat64(row.BasalArea)
		volume := utils.NullStringToNullFloat64(row.Volume)

		specimen := &types.SpecimenWithSpecies{
			ID:              row.SpecimenID.String,
			Portion:         row.Portion.String,
			Height:          *height,
			Cap1:            *cap1,
			Cap2:            utils.NullStringToNullFloat64(row.Cap2),
			Cap3:            utils.NullStringToNullFloat64(row.Cap3),
			Cap4:            utils.NullStringToNullFloat64(row.Cap4),
			Cap5:            utils.NullStringToNullFloat64(row.Cap5),
			Cap6:            utils.NullStringToNullFloat64(row.Cap6),
			AverageDap:      *averageDap,
			BasalArea:       *basalArea,
			Volume:          *volume,
			RegisterDate:    row.RegisterDate.Time,
			PhytoAnalysisID: firstRow.PhytoID,
			SpecieID:        row.SpecieID.String,
			ScientificName:  row.ScientificName.String,
			Family:          row.Family.String,
			PopularName:     utils.FromNullString(row.PopularName),
		}

		result.Specimens = append(result.Specimens, specimen)
	}

	return result, nil
}
