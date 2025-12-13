package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainparameter "github.com/ESG-Project/suassu-api/internal/domain/parameter"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type ParameterRepo struct{ q *sqlc.Queries }

func NewParameterRepoFrom(d dbtx) *ParameterRepo {
	return &ParameterRepo{q: sqlc.New(d)}
}

func NewParameterRepo(db *sql.DB) *ParameterRepo {
	return &ParameterRepo{q: sqlc.New(db)}
}

func (r *ParameterRepo) Create(ctx context.Context, p *domainparameter.Parameter) error {
	return r.q.CreateParameter(ctx, sqlc.CreateParameterParams{
		ID:           p.ID,
		Title:        p.Title,
		Value:        utils.ToNullString(p.Value),
		EnterpriseId: p.EnterpriseID,
		IsDefault:    p.IsDefault,
	})
}

func (r *ParameterRepo) GetByID(ctx context.Context, id, enterpriseID string) (*domainparameter.Parameter, error) {
	row, err := r.q.GetParameterByID(ctx, sqlc.GetParameterByIDParams{
		ID:           id,
		EnterpriseId: enterpriseID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "parameter not found")
		}
		return nil, err
	}

	return &domainparameter.Parameter{
		ID:           row.ID,
		Title:        row.Title,
		Value:        utils.FromNullString(row.Value),
		EnterpriseID: row.EnterpriseId,
		IsDefault:    row.IsDefault,
	}, nil
}

func (r *ParameterRepo) List(ctx context.Context, enterpriseID string) ([]domainparameter.Parameter, error) {
	rows, err := r.q.ListParametersByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	params := make([]domainparameter.Parameter, 0, len(rows))
	for _, row := range rows {
		params = append(params, domainparameter.Parameter{
			ID:           row.ID,
			Title:        row.Title,
			Value:        utils.FromNullString(row.Value),
			EnterpriseID: row.EnterpriseId,
			IsDefault:    row.IsDefault,
		})
	}
	return params, nil
}

func (r *ParameterRepo) Update(ctx context.Context, p *domainparameter.Parameter) error {
	return r.q.UpdateParameter(ctx, sqlc.UpdateParameterParams{
		ID:           p.ID,
		Title:        p.Title,
		Value:        utils.ToNullString(p.Value),
		IsDefault:    p.IsDefault,
		EnterpriseId: p.EnterpriseID,
	})
}

func (r *ParameterRepo) Delete(ctx context.Context, id, enterpriseID string) error {
	return r.q.DeleteParameter(ctx, sqlc.DeleteParameterParams{
		ID:           id,
		EnterpriseId: enterpriseID,
	})
}
