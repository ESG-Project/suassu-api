package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
	domainenterprise "github.com/ESG-Project/suassu-api/internal/domain/enterprise"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type UserRepo struct {
	q   *sqlc.Queries
	txm TxManagerInterface
}

// Construtor que aceita qualquer sqlc.DBTX (ex.: *sql.DB ou *sql.Tx)
func NewUserRepoFrom(d dbtx) *UserRepo {
	return &UserRepo{q: sqlc.New(d)}
}

// Compatibilidade com construtor anterior
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{q: sqlc.New(db)}
}

// Construtor com TxManager para operações transacionais
func NewUserRepoWithTx(db *sql.DB, txm TxManagerInterface) *UserRepo {
	return &UserRepo{q: sqlc.New(db), txm: txm}
}

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
			addr := &domainaddress.Address{ID: row.AddressID.String}
			if row.ZipCode.Valid {
				addr.ZipCode = row.ZipCode.String
			}
			if row.State.Valid {
				addr.State = row.State.String
			}
			if row.City.Valid {
				addr.City = row.City.String
			}
			if row.Neighborhood.Valid {
				addr.Neighborhood = row.Neighborhood.String
			}
			if row.Street.Valid {
				addr.Street = row.Street.String
			}
			if row.Num.Valid {
				addr.Num = row.Num.String
			}
			if row.Latitude.Valid {
				addr.SetLatitude(&row.Latitude.String)
			}
			if row.Longitude.Valid {
				addr.SetLongitude(&row.Longitude.String)
			}
			if row.AddInfo.Valid {
				addr.SetAddInfo(&row.AddInfo.String)
			}
			u.SetAddress(addr)
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
	row, err := r.q.FindUserByEmailForAuth(ctx, email)
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

// GetByIDForRefresh - busca usuário por ID sem filtro de tenant (para refresh token)
func (r *UserRepo) GetByIDForRefresh(ctx context.Context, userID string) (*domainuser.User, error) {
	row, err := r.q.GetUserByIDForRefresh(ctx, userID)
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
	// Usar a query GetUserByID com filtro de email
	// Por enquanto, vamos implementar uma busca simples
	users, _, err := r.List(ctx, enterpriseID, 1000, nil)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, apperr.New(apperr.CodeNotFound, "user not found")
}

// GetUserPermissionsWithRole - usando transação para buscar dados relacionados
func (r *UserRepo) GetUserPermissionsWithRole(ctx context.Context, userID string, enterpriseID string) (*types.UserPermissions, error) {
	if r.txm == nil {
		return nil, apperr.New(apperr.CodeInvalid, "transaction manager required")
	}

	var result *types.UserPermissions

	err := r.txm.RunInTx(ctx, func(repos Repos) error {
		// 1. Buscar usuário básico
		userRow, err := r.q.GetUserByID(ctx, sqlc.GetUserByIDParams{
			EnterpriseId: enterpriseID,
			ID:           userID,
		})
		if err != nil {
			return err
		}

		// 2. Buscar role (se existir)
		var roleTitle string
		if userRow.RoleID.Valid {
			role, err := repos.Roles().GetByID(ctx, userRow.RoleID.String, enterpriseID)
			if err != nil {
				// Log do erro mas não falha a query principal
				roleTitle = ""
			} else {
				roleTitle = role.Title
			}
		}

		// 3. Buscar permissões (se houver role)
		var permissions []*types.UserPermission
		if userRow.RoleID.Valid {
			if permissions, err = repos.Permissions().GetByRoleID(ctx, userRow.RoleID.String); err != nil {
				// Log do erro mas não falha a query principal
				permissions = []*types.UserPermission{}
			}
		}

		result = &types.UserPermissions{
			ID:          userRow.ID,
			Name:        userRow.Name,
			RoleTitle:   roleTitle,
			Permissions: permissions,
		}

		return nil
	})

	return result, err
}

// GetUserWithDetails - usando transação para buscar dados relacionados
func (r *UserRepo) GetUserWithDetails(ctx context.Context, userID string, enterpriseID string) (*types.UserWithDetails, error) {
	if r.txm == nil {
		return nil, apperr.New(apperr.CodeInvalid, "transaction manager required")
	}

	var result *types.UserWithDetails

	err := r.txm.RunInTx(ctx, func(repos Repos) error {
		// 1. Buscar usuário básico
		userRow, err := r.q.GetUserByID(ctx, sqlc.GetUserByIDParams{
			EnterpriseId: enterpriseID,
			ID:           userID,
		})
		if err != nil {
			return err
		}

		// 2. Buscar role (se existir)
		var role *types.UserRole
		if userRow.RoleID.Valid {
			if role, err = repos.Roles().GetByID(ctx, userRow.RoleID.String, enterpriseID); err != nil {
				// Log do erro mas não falha a query principal
				role = nil
			}
		}

		// 3. Buscar empresa
		var domainEnterprise *domainenterprise.Enterprise
		if domainEnterprise, err = repos.Enterprises().GetByID(ctx, userRow.EnterpriseID); err != nil {
			return err
		}

		// Converter domain.Enterprise para types.UserEnterprise
		var enterpriseAddress *types.UserAddress
		if domainEnterprise.Address != nil {
			enterpriseAddress = &types.UserAddress{
				ID:           domainEnterprise.Address.ID,
				ZipCode:      domainEnterprise.Address.ZipCode,
				State:        domainEnterprise.Address.State,
				City:         domainEnterprise.Address.City,
				Neighborhood: domainEnterprise.Address.Neighborhood,
				Street:       domainEnterprise.Address.Street,
				Num:          domainEnterprise.Address.Num,
				Latitude:     domainEnterprise.Address.Latitude,
				Longitude:    domainEnterprise.Address.Longitude,
				AddInfo:      domainEnterprise.Address.AddInfo,
			}
		}

		enterprise := &types.UserEnterprise{
			ID:          domainEnterprise.ID,
			Name:        domainEnterprise.Name,
			CNPJ:        domainEnterprise.CNPJ,
			Email:       domainEnterprise.Email,
			FantasyName: domainEnterprise.FantasyName,
			Phone:       domainEnterprise.Phone,
			AddressID:   domainEnterprise.AddressID,
			Address:     enterpriseAddress,
		}

		// 4. Buscar endereço (se existir)
		var address *types.UserAddress
		if userRow.AddressID.Valid {
			if address, err = repos.Addresses().GetByID(ctx, userRow.AddressID.String); err != nil {
				// Log do erro mas não falha a query principal
				// address permanece nil
			}
		}

		// 5. Buscar permissões (se houver role)
		var permissions []*types.UserPermission
		if role != nil {
			if permissions, err = repos.Permissions().GetByRoleID(ctx, role.ID); err != nil {
				// Log do erro mas não falha a query principal
				permissions = []*types.UserPermission{}
			}
		}

		result = &types.UserWithDetails{
			ID:           userRow.ID,
			Name:         userRow.Name,
			Email:        userRow.Email,
			Document:     userRow.Document,
			Phone:        utils.FromNullString(userRow.Phone),
			EnterpriseID: userRow.EnterpriseID,
			Role:         role,
			Enterprise:   enterprise,
			Address:      address,
			Permissions:  permissions,
		}

		return nil
	})

	return result, err
}
