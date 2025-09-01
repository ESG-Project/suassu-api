package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type PermissionRepo struct{ q *sqlc.Queries }

func NewPermissionRepoFrom(d dbtx) *PermissionRepo {
	return &PermissionRepo{q: sqlc.New(d)}
}

func NewPermissionRepo(db *sql.DB) *PermissionRepo {
	return &PermissionRepo{q: sqlc.New(db)}
}

func (r *PermissionRepo) GetByRoleID(ctx context.Context, roleID string) ([]*types.UserPermission, error) {
	rows, err := r.q.ListPermissionsByRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	permissions := make([]*types.UserPermission, len(rows))
	for i, row := range rows {
		permissions[i] = &types.UserPermission{
			ID:          row.ID,
			FeatureID:   row.FeatureID,
			FeatureName: row.FeatureName,
			Create:      row.Create,
			Read:        row.Read,
			Update:      row.Update,
			Delete:      row.Delete,
		}
	}

	return permissions, nil
}
