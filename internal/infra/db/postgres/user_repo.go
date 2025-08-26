package postgres

import (
	"context"
	"database/sql"

	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
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

func (r *UserRepo) Create(ctx context.Context, u *domainuser.User) error {
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

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	user := domainuser.NewUser(
		row.ID,
		row.Name,
		row.Email,
		row.PasswordHash, // alias do SELECT
		row.Document,
		row.EnterpriseID,
	)

	// Set optional fields
	if row.Phone.Valid {
		user.SetPhone(&row.Phone.String)
	}
	if row.AddressID.Valid {
		user.SetAddressID(&row.AddressID.String)
	}
	if row.RoleID.Valid {
		user.SetRoleID(&row.RoleID.String)
	}

	return user, nil
}

func (r *UserRepo) List(ctx context.Context, limit, offset int32) ([]*domainuser.User, error) {
	rows, err := r.q.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*domainuser.User, 0, len(rows))
	for _, row := range rows {
		user := domainuser.NewUser(
			row.ID,
			row.Name,
			row.Email,
			row.PasswordHash,
			row.Document,
			row.EnterpriseID,
		)

		// Set optional fields
		if row.Phone.Valid {
			user.SetPhone(&row.Phone.String)
		}
		if row.AddressID.Valid {
			user.SetAddressID(&row.AddressID.String)
		}
		if row.RoleID.Valid {
			user.SetRoleID(&row.RoleID.String)
		}

		out = append(out, user)
	}
	return out, nil
}
