package role

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainrole "github.com/ESG-Project/suassu-api/internal/domain/role"
)

// Repo define a interface do repositório de Role.
type Repo interface {
	Create(ctx context.Context, r *domainrole.Role) error
	List(ctx context.Context, enterpriseID string) ([]domainrole.Role, error)
	GetByID(ctx context.Context, roleID, enterpriseID string) (*types.UserRole, error)
	Delete(ctx context.Context, roleID string) error
}

// PermissionsByRoleGetter resolve a matriz de permissões de um papel específico,
// usada para comparar o cargo-alvo contra o ator em ListAssignable.
type PermissionsByRoleGetter interface {
	GetByRoleID(ctx context.Context, roleID string) ([]*types.UserPermission, error)
}

// UserDetails resolve o papel e as permissões do usuário logado, usadas para
// calcular quais cargos ele pode atribuir a outro usuário.
type UserDetails interface {
	GetUserWithDetails(ctx context.Context, userID, enterpriseID string) (*types.UserWithDetails, error)
	GetUserPermissionsWithRole(ctx context.Context, userID, enterpriseID string) (*types.UserPermissions, error)
}
