package adminuser

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainclient "github.com/ESG-Project/suassu-api/internal/domain/client"
	domaintechnician "github.com/ESG-Project/suassu-api/internal/domain/technician"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	postgres "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
)

// UserRepo é o subconjunto de operações administrativas sobre User usadas por
// este serviço — satisfeito estruturalmente por *postgres.UserRepo.
type UserRepo interface {
	GetByID(ctx context.Context, userID, enterpriseID string) (*domainuser.User, error)
	GetByEmailForAuth(ctx context.Context, email string) (*domainuser.User, error)
	GetByDocumentInEnterprise(ctx context.Context, enterpriseID, document string) (*domainuser.User, error)
	GetPrimaryAdminUserID(ctx context.Context, enterpriseID string) (string, error)
	Delete(ctx context.Context, userID, enterpriseID string) error
	ListNonClientUsersByEnterprise(ctx context.Context, enterpriseID, excludeID string) ([]postgres.NonClientUserRow, error)
}

// RoleRepo resolve papéis por id ou por título (aceito historicamente pelo
// user-crud em POST/PUT /user).
type RoleRepo interface {
	GetByID(ctx context.Context, roleID, enterpriseID string) (*types.UserRole, error)
	GetByTitle(ctx context.Context, enterpriseID, title string) (*types.UserRole, error)
}

// PermissionRepo resolve a matriz de permissões de um papel específico, usada
// para a checagem de anti-escalação contra o ator.
type PermissionRepo interface {
	GetByRoleID(ctx context.Context, roleID string) ([]*types.UserPermission, error)
}

// ClientRepo é o subconjunto não-transacional de operações sobre Client.
type ClientRepo interface {
	GetByUserID(ctx context.Context, userID string) (*domainclient.Client, error)
	ListByEnterprise(ctx context.Context, enterpriseID string) ([]postgres.ClientListRow, error)
}

// TechnicianRepo é o subconjunto não-transacional de operações sobre Technician.
type TechnicianRepo interface {
	GetByUserID(ctx context.Context, userID string) (*domaintechnician.Technician, error)
}

// ActorPermissions resolve o papel e a matriz de permissões do usuário
// logado, para a checagem de anti-escalação.
type ActorPermissions interface {
	GetUserWithDetails(ctx context.Context, userID, enterpriseID string) (*types.UserWithDetails, error)
	GetUserPermissionsWithRole(ctx context.Context, userID, enterpriseID string) (*types.UserPermissions, error)
}

// UserDetails resolve um usuário com todos os detalhes (endereço, papel,
// empresa) — reaproveitado de appuser.Service para a rota GET /user/:id.
type UserDetails interface {
	GetUserWithDetails(ctx context.Context, userID, enterpriseID string) (*types.UserWithDetails, error)
}
