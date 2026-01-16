package user

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
)

type Repo interface {
	Create(ctx context.Context, u *domainuser.User) error
	List(ctx context.Context, enterpriseID string, limit int32, after *domainuser.UserCursorKey) ([]*domainuser.User, domainuser.PageInfo, error)
	GetByEmailForAuth(ctx context.Context, email string) (*domainuser.User, error)    // Para autenticação (sem filtro de tenant)
	GetByIDForRefresh(ctx context.Context, userID string) (*domainuser.User, error)   // Para refresh token (sem filtro de tenant)
	GetUserWithDetails(ctx context.Context, userID string, enterpriseID string) (*types.UserWithDetails, error)
	GetUserPermissionsWithRole(ctx context.Context, userID string, enterpriseID string) (*types.UserPermissions, error)
}

type Hasher interface {
	Hash(pw string) (string, error)
	Compare(hash, plain string) error
}
