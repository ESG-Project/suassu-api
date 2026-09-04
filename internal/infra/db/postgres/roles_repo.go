package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainrole "github.com/ESG-Project/suassu-api/internal/domain/role"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type RoleRepo struct{ q *sqlc.Queries }

func NewRoleRepoFrom(d dbtx) *RoleRepo {
	return &RoleRepo{q: sqlc.New(d)}
}

func NewRoleRepo(db *sql.DB) *RoleRepo {
	return &RoleRepo{q: sqlc.New(db)}
}

func (r *RoleRepo) Create(ctx context.Context, role *domainrole.Role) error {
	_, err := r.q.CreateRole(ctx, sqlc.CreateRoleParams{
		ID:           role.ID,
		Title:        role.Title,
		EnterpriseId: role.EnterpriseID,
	})
	return err
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

// GetByTitle resolve um papel pelo título dentro da empresa (ex.: "Técnico",
// "Cliente"). Usado quando o cliente informa o título em vez do UUID do
// papel (aceito historicamente pelo user-crud em POST/PUT /user).
func (r *RoleRepo) GetByTitle(ctx context.Context, enterpriseID, title string) (*types.UserRole, error) {
	row, err := r.q.GetRoleByTitle(ctx, sqlc.GetRoleByTitleParams{
		EnterpriseId: enterpriseID,
		Title:        title,
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

func (r *RoleRepo) Delete(ctx context.Context, roleID string) error {
	return r.q.DeleteRole(ctx, roleID)
}

func (r *RoleRepo) List(ctx context.Context, enterpriseID string) ([]domainrole.Role, error) {
	rows, err := r.q.ListRolesByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	roles := make([]domainrole.Role, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, domainrole.Role{
			ID:           row.ID,
			Title:        row.Title,
			EnterpriseID: row.EnterpriseID,
		})
	}
	return roles, nil
}
