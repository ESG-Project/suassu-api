package postgres

import (
	"context"
	"database/sql"

	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
	"github.com/google/uuid"
)

type UserCursorKey struct {
	Email string    `json:"email"`
	ID    uuid.UUID `json:"id"`
}

type PageInfo struct {
	Next    *UserCursorKey
	HasMore bool
}

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
		Password:     u.PasswordHash,
		Document:     u.Document,
		Phone:        toNullString(u.Phone),
		AddressId:    toNullString(u.AddressID),
		RoleId:       toNullString(u.RoleID),
		EnterpriseId: u.EnterpriseID,
	})
}

func (r *UserRepo) List(ctx context.Context, enterpriseID string, limit int32, after *UserCursorKey) ([]*domainuser.User, PageInfo, error) {
	params := sqlc.ListUsersParams{
		EnterpriseId: enterpriseID,
		Limit:        limit + 1,
	}

	// Se não há cursor, usar valores vazios para pegar todos
	if after == nil {
		params.Email = ""
		params.ID = ""
	} else {
		params.Email = after.Email
		params.ID = after.ID.String()
	}

	rows, err := r.q.ListUsers(ctx, params)
	if err != nil {
		return nil, PageInfo{}, err
	}

	hasMore := false
	if int32(len(rows)) > limit {
		hasMore = true
		rows = rows[:limit]
	}

	out := make([]*domainuser.User, 0, len(rows))
	for _, row := range rows {
		u := domainuser.NewUser(
			row.ID,
			row.Name,
			row.Email,
			row.PasswordHash,
			row.Document,
			row.EnterpriseID,
		)
		// opcionais
		if row.Phone.Valid {
			u.SetPhone(&row.Phone.String)
		}
		if row.AddressID.Valid {
			u.SetAddressID(&row.AddressID.String)
		}
		if row.RoleID.Valid {
			u.SetRoleID(&row.RoleID.String)
		}
		out = append(out, u)
	}

	var next *UserCursorKey
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		next = &UserCursorKey{Email: last.Email}
	}

	return out, PageInfo{Next: next, HasMore: hasMore}, nil
}

// GetByEmailForAuth - específico para autenticação (sem filtro de tenant)
func (r *UserRepo) GetByEmailForAuth(ctx context.Context, email string) (*domainuser.User, error) {
	row, err := r.q.GetUserByEmailForAuth(ctx, email)
	if err != nil {
		return nil, err
	}

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

	return user, nil
}

// GetByEmailInTenant - para operações de negócio (com filtro de tenant)
func (r *UserRepo) GetByEmailInTenant(ctx context.Context, email, enterpriseID string) (*domainuser.User, error) {
	row, err := r.q.GetUserByEmailInTenant(ctx, sqlc.GetUserByEmailInTenantParams{
		EnterpriseId: enterpriseID,
		Email:        email,
	})
	if err != nil {
		return nil, err
	}

	user := domainuser.NewUser(
		row.ID,
		row.Name,
		row.Email,
		row.PasswordHash,
		row.Document,
		row.EnterpriseID,
	)

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
