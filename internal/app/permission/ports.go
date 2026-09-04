package permission

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainfeature "github.com/ESG-Project/suassu-api/internal/domain/feature"
	domainpermission "github.com/ESG-Project/suassu-api/internal/domain/permission"
)

// Repo define a interface do repositório de Permission.
type Repo interface {
	Create(ctx context.Context, p *domainpermission.Permission) error
	Update(ctx context.Context, p *domainpermission.Permission) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*types.PermissionWithEnterprise, error)
	ListByEnterprise(ctx context.Context, enterpriseID string) ([]domainpermission.Permission, error)
}

// RoleGetter confirma que um papel existe e pertence à empresa do ator.
type RoleGetter interface {
	GetByID(ctx context.Context, roleID, enterpriseID string) (*types.UserRole, error)
}

// FeatureGetter confirma que uma feature existe.
type FeatureGetter interface {
	GetByID(ctx context.Context, id string) (*domainfeature.Feature, error)
}

// ActorPermissions resolve a matriz de permissões do usuário logado, usada
// para a checagem de anti-escalação (só pode conceder o que já possui).
type ActorPermissions interface {
	GetUserPermissionsWithRole(ctx context.Context, userID, enterpriseID string) (*types.UserPermissions, error)
}
