package enterprise

import (
	"context"

	domainenterprise "github.com/ESG-Project/suassu-api/internal/domain/enterprise"
)

type Repo interface {
	Create(ctx context.Context, e *domainenterprise.Enterprise) error
	GetByID(ctx context.Context, id string) (*domainenterprise.Enterprise, error)
	Update(ctx context.Context, e *domainenterprise.Enterprise) error
}

type Hasher interface {
	Hash(pw string) (string, error)
	Compare(hash, plain string) error
}
