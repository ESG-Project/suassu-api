package user

import (
	"context"

	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
)

type Repo interface {
	Create(ctx context.Context, u *domainuser.User) error
	GetByEmail(ctx context.Context, email string) (*domainuser.User, error)
	List(ctx context.Context, enterpriseID string, limit, offset int32) ([]*domainuser.User, error)
}

type Hasher interface {
	Hash(pw string) (string, error)
	Compare(hash, plain string) error
}
