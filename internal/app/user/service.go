package user

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
)

type ServiceInterface interface {
	Create(ctx context.Context, enterpriseID string, in CreateInput) (string, error)
	List(ctx context.Context, enterpriseID string, limit int32, after *domainuser.UserCursorKey) ([]domainuser.User, *domainuser.PageInfo, error)
	GetUserWithDetails(ctx context.Context, userID string, enterpriseID string) (*types.UserWithDetails, error)
	GetUserPermissionsWithRole(ctx context.Context, userID string, enterpriseID string) (*types.UserPermissions, error)
}
