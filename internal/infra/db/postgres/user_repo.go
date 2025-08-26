package postgres

import (
	"context"
	"database/sql"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}
func fromNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	str := ns.String
	return &str
}

type UserRepo struct{ q *sqlc.Queries }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{q: sqlc.New(db)} }

func (r *UserRepo) Create(ctx context.Context, u appuser.Entity) error {
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		Password:     u.PasswordHash, // coluna "password"
		Document:     u.Document,
		Phone:        toNullString(u.Phone),
		AddressId:    toNullString(u.AddressID),
		RoleId:       toNullString(u.RoleID),
		EnterpriseId: u.EnterpriseID,
	})
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (appuser.Entity, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return appuser.Entity{}, err
	}
	return appuser.Entity{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: row.PasswordHash, // alias do SELECT
		Document:     row.Document,
		Phone:        fromNullString(row.Phone),
		AddressID:    fromNullString(row.AddressID),
		RoleID:       fromNullString(row.RoleID),
		EnterpriseID: row.EnterpriseID,
	}, nil
}

func (r *UserRepo) List(ctx context.Context, limit, offset int32) ([]appuser.Entity, error) {
	rows, err := r.q.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]appuser.Entity, 0, len(rows))
	for _, row := range rows {
		out = append(out, appuser.Entity{
			ID:           row.ID,
			Name:         row.Name,
			Email:        row.Email,
			PasswordHash: row.PasswordHash,
			Document:     row.Document,
			Phone:        fromNullString(row.Phone),
			AddressID:    fromNullString(row.AddressID),
			RoleID:       fromNullString(row.RoleID),
			EnterpriseID: row.EnterpriseID,
		})
	}
	return out, nil
}
