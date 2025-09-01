package user

import (
	"context"

	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/http/dto"
)

type ServiceInterface interface {
	Create(ctx context.Context, enterpriseID string, in CreateInput) (string, error)
	List(ctx context.Context, enterpriseID string, limit int32, after *domainuser.UserCursorKey) ([]domainuser.User, *domainuser.PageInfo, error)
	GetByEmailInTenant(ctx context.Context, enterpriseID string, email string) (*domainuser.User, error)
	GetUserWithDetails(ctx context.Context, userID string, enterpriseID string) (*dto.MeOut, error)
	GetUserPermissionsWithRole(ctx context.Context, userID string, enterpriseID string) (*dto.MyPermissionsOut, error)
}
