package user

import (
	"context"

	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/http/dto"
)

type Repo interface {
	Create(ctx context.Context, u *domainuser.User) error
	List(ctx context.Context, enterpriseID string, limit int32, after *domainuser.UserCursorKey) ([]*domainuser.User, domainuser.PageInfo, error)
	GetByEmailInTenant(ctx context.Context, enterpriseID string, email string) (*domainuser.User, error) // Para operações de negócio (com filtro de tenant)
	GetByEmailForAuth(ctx context.Context, email string) (*domainuser.User, error)                       // Para autenticação (sem filtro de tenant)
	GetUserWithDetails(ctx context.Context, userID string, enterpriseID string) (*dto.MeOut, error)
	GetUserPermissionsWithRole(ctx context.Context, userID string, enterpriseID string) (*dto.MyPermissionsOut, error)
}

type Hasher interface {
	Hash(pw string) (string, error)
	Compare(hash, plain string) error
}
