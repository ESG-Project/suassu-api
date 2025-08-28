package postgres

import (
	"context"
	"database/sql"

	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
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
	var next *UserCursorKey

	if int32(len(rows)) > limit {
		hasMore = true
		// Calcular next ANTES de truncar
		last := rows[limit-1] // Usar índice limit-1, não len(rows)-1
		lastID, _ := uuid.Parse(last.ID)
		next = &UserCursorKey{Email: last.Email, ID: lastID}
		rows = rows[:limit]
	} else if len(rows) > 0 {
		// Se não há mais páginas, next é nil
		last := rows[len(rows)-1]
		lastID, _ := uuid.Parse(last.ID)
		next = &UserCursorKey{Email: last.Email, ID: lastID}
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
			u.SetAddress(&domainaddress.Address{
				ID:           row.AddressID.String,
				ZipCode:      row.ZipCode,
				Street:       row.Street,
				City:         row.City,
				State:        row.State,
				Neighborhood: row.Neighborhood,
				Num:          row.Num,
			})
			if row.Latitude.Valid {
				u.Address.SetLatitude(&row.Latitude.String)
			}
			if row.Longitude.Valid {
				u.Address.SetLongitude(&row.Longitude.String)
			}
			if row.AddInfo.Valid {
				u.Address.SetAddInfo(&row.AddInfo.String)
			}
		}
		if row.RoleID.Valid {
			u.SetRoleID(&row.RoleID.String)
		}
		out = append(out, u)
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
