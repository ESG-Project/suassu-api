package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type RoleRepo struct{ q *sqlc.Queries }

func NewRoleRepoFrom(d dbtx) *RoleRepo {
	return &RoleRepo{q: sqlc.New(d)}
}

func NewRoleRepo(db *sql.DB) *RoleRepo {
	return &RoleRepo{q: sqlc.New(db)}
}

func (r *RoleRepo) GetByID(ctx context.Context, roleID string, enterpriseID string) (*types.UserRole, error) {
	row, err := r.q.GetRoleByID(ctx, sqlc.GetRoleByIDParams{
		ID:           roleID,
		EnterpriseId: enterpriseID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "role not found")
		}
		return nil, err
	}

	return &types.UserRole{
		ID:    row.ID,
		Title: row.Title,
	}, nil
}
