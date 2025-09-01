package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type UserRepo struct{ q *sqlc.Queries }

// Construtor que aceita qualquer sqlc.DBTX (ex.: *sql.DB ou *sql.Tx)
func NewUserRepoFrom(d dbtx) *UserRepo { return &UserRepo{q: sqlc.New(d)} }

// Compatibilidade com construtor anterior
func NewUserRepo(db *sql.DB) *UserRepo { return NewUserRepoFrom(db) }

func (r *UserRepo) Create(ctx context.Context, u *domainuser.User) error {
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		Password:     u.PasswordHash,
		Document:     u.Document,
		Phone:        utils.ToNullString(u.Phone),
		AddressId:    utils.ToNullString(u.AddressID),
		RoleId:       utils.ToNullString(u.RoleID),
		EnterpriseId: u.EnterpriseID,
	})
}

func (r *UserRepo) List(ctx context.Context, enterpriseID string, limit int32, after *domainuser.UserCursorKey) ([]*domainuser.User, domainuser.PageInfo, error) {
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
		return nil, domainuser.PageInfo{}, err
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

	return out, domainuser.PageInfo{}, nil
}

// GetByEmailForAuth - específico para autenticação (sem filtro de tenant)
func (r *UserRepo) GetByEmailForAuth(ctx context.Context, email string) (*domainuser.User, error) {
	row, err := r.q.GetUserByEmailForAuth(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "user not found")
		}
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
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "user not found")
		}
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

func (r *UserRepo) GetUserPermissionsWithRole(ctx context.Context, userID string, enterpriseID string) (*types.UserPermissions, error) {
	row, err := r.q.GetUserPermissionsWithRole(ctx, sqlc.GetUserPermissionsWithRoleParams{
		ID:           userID,
		EnterpriseId: enterpriseID,
	})
	if err != nil {
		return nil, err
	}

	var permissions []*types.UserPermission
	if row.Permissions != nil {
		type permissionJSON struct {
			ID          string `json:"id"`
			FeatureID   string `json:"feature_id"`
			FeatureName string `json:"feature_name"`
			Create      bool   `json:"create"`
			Read        bool   `json:"read"`
			Update      bool   `json:"update"`
			Delete      bool   `json:"delete"`
		}

		var permJSON []permissionJSON

		if permBytes, ok := row.Permissions.([]uint8); ok {
			if err := json.Unmarshal(permBytes, &permJSON); err == nil {
				permissions = make([]*types.UserPermission, len(permJSON))
				for i, p := range permJSON {
					permissions[i] = &types.UserPermission{
						ID:          p.ID,
						FeatureID:   p.FeatureID,
						FeatureName: p.FeatureName,
						Create:      p.Create,
						Read:        p.Read,
						Update:      p.Update,
						Delete:      p.Delete,
					}
				}
			}
		}
	}

	return &types.UserPermissions{
		ID:          row.UserID,
		Name:        row.UserName,
		RoleTitle:   row.RoleTitle,
		Permissions: permissions,
	}, nil
}

func (r *UserRepo) GetUserWithDetails(ctx context.Context, userID string, enterpriseID string) (*types.UserWithDetails, error) {
	row, err := r.q.GetUserWithDetails(ctx, sqlc.GetUserWithDetailsParams{
		ID:           userID,
		EnterpriseId: enterpriseID,
	})
	if err != nil {
		return nil, err
	}

	// Converter permissões do JSON para slice
	var permissions []*types.UserPermission
	if row.Permissions != nil {
		// Estrutura auxiliar para unmarshal
		type permissionJSON struct {
			ID          string `json:"id"`
			FeatureID   string `json:"feature_id"`
			FeatureName string `json:"feature_name"`
			Create      bool   `json:"create"`
			Read        bool   `json:"read"`
			Update      bool   `json:"update"`
			Delete      bool   `json:"delete"`
		}

		var permJSON []permissionJSON

		if permBytes, ok := row.Permissions.([]uint8); ok {
			if err := json.Unmarshal(permBytes, &permJSON); err == nil {
				permissions = make([]*types.UserPermission, len(permJSON))
				for i, p := range permJSON {
					permissions[i] = &types.UserPermission{
						ID:          p.ID,
						FeatureID:   p.FeatureID,
						FeatureName: p.FeatureName,
						Create:      p.Create,
						Read:        p.Read,
						Update:      p.Update,
						Delete:      p.Delete,
					}
				}
			}
		}
	}

	// Construir endereço se existir
	var address *types.UserAddress
	if row.AddressID.Valid {
		address = &types.UserAddress{
			ID:           row.AddressID.String,
			ZipCode:      row.AddressZipCode.String,
			State:        row.AddressState.String,
			City:         row.AddressCity.String,
			Neighborhood: row.AddressNeighborhood.String,
			Street:       row.AddressStreet.String,
			Num:          row.AddressNum.String,
		}
		if row.AddressLatitude.Valid {
			address.Latitude = &row.AddressLatitude.String
		}
		if row.AddressLongitude.Valid {
			address.Longitude = &row.AddressLongitude.String
		}
		if row.AddressAddInfo.Valid {
			address.AddInfo = &row.AddressAddInfo.String
		}
	}

	// Construir empresa
	enterprise := &types.UserEnterprise{
		ID:    row.EnterpriseID,
		Name:  row.EnterpriseName,
		CNPJ:  row.EnterpriseCnpj,
		Email: row.EnterpriseEmail,
	}
	if row.EnterpriseFantasyName.Valid {
		enterprise.FantasyName = &row.EnterpriseFantasyName.String
	}
	if row.EnterprisePhone.Valid {
		enterprise.Phone = &row.EnterprisePhone.String
	}
	if row.EnterpriseAddressID.Valid {
		enterprise.AddressID = &row.EnterpriseAddressID.String
	}

	// Construir role
	var role *types.UserRole
	if row.UserRoleID.Valid {
		role = &types.UserRole{
			ID:    row.UserRoleID.String,
			Title: row.RoleTitle,
		}
	}

	// Construir resposta final
	return &types.UserWithDetails{
		ID:           row.UserID,
		Name:         row.UserName,
		Email:        row.UserEmail,
		Document:     row.UserDocument,
		Phone:        utils.FromNullString(row.UserPhone),
		Address:      address,
		Role:         role,
		EnterpriseID: row.UserEnterpriseID,
		Enterprise:   enterprise,
		Permissions:  permissions,
	}, nil
}
