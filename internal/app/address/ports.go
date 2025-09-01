package address

import (
	"context"

	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
)

type Hasher interface {
	Hash(pw string) (string, error)
	Compare(hash, plain string) error
}

type Repo interface {
	Create(ctx context.Context, a *domainaddress.Address) error
	FindByDetails(ctx context.Context, params *domainaddress.SearchParams) (*domainaddress.Address, error)
}
