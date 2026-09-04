package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainpermission "github.com/ESG-Project/suassu-api/internal/domain/permission"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type PermissionRepo struct{ q *sqlc.Queries }

func NewPermissionRepoFrom(d dbtx) *PermissionRepo {
	return &PermissionRepo{q: sqlc.New(d)}
}

func NewPermissionRepo(db *sql.DB) *PermissionRepo {
	return &PermissionRepo{q: sqlc.New(db)}
}

func (r *PermissionRepo) Create(ctx context.Context, permission *domainpermission.Permission) error {
	_, err := r.q.CreatePermission(ctx, sqlc.CreatePermissionParams{
		ID:        permission.ID,
		FeatureId: permission.FeatureID,
		RoleId:    permission.RoleID,
		Create:    permission.Create,
		Read:      permission.Read,
		Update:    permission.Update,
		Delete:    permission.Delete,
	})
	return err
}

func (r *PermissionRepo) GetByID(ctx context.Context, id string) (*types.PermissionWithEnterprise, error) {
	row, err := r.q.GetPermissionByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "permission not found")
		}
		return nil, err
	}

	return &types.PermissionWithEnterprise{
		ID:           row.ID,
		FeatureID:    row.FeatureID,
		RoleID:       row.RoleID,
		Create:       row.Create,
		Read:         row.Read,
		Update:       row.Update,
		Delete:       row.Delete,
		EnterpriseID: row.EnterpriseID,
	}, nil
}

func (r *PermissionRepo) ListByEnterprise(ctx context.Context, enterpriseID string) ([]domainpermission.Permission, error) {
	rows, err := r.q.ListPermissionsByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	permissions := make([]domainpermission.Permission, 0, len(rows))
	for _, row := range rows {
		permissions = append(permissions, domainpermission.Permission{
			ID:        row.ID,
			FeatureID: row.FeatureID,
			RoleID:    row.RoleID,
			Create:    row.Create,
			Read:      row.Read,
			Update:    row.Update,
			Delete:    row.Delete,
		})
	}
	return permissions, nil
}

func (r *PermissionRepo) Update(ctx context.Context, permission *domainpermission.Permission) error {
	_, err := r.q.UpdatePermission(ctx, sqlc.UpdatePermissionParams{
		ID:        permission.ID,
		FeatureId: permission.FeatureID,
		RoleId:    permission.RoleID,
		Create:    permission.Create,
		Read:      permission.Read,
		Update:    permission.Update,
		Delete:    permission.Delete,
	})
	return err
}

func (r *PermissionRepo) Delete(ctx context.Context, id string) error {
	return r.q.DeletePermission(ctx, id)
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
