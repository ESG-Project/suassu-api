package feature

import (
	"context"

	domainfeature "github.com/ESG-Project/suassu-api/internal/domain/feature"
)

type Repo interface {
	Upsert(ctx context.Context, name string) error
	List(ctx context.Context) ([]domainfeature.Feature, error)
	GetByID(ctx context.Context, id string) (*domainfeature.Feature, error)
	Create(ctx context.Context, f *domainfeature.Feature) error
	Delete(ctx context.Context, id string) error
}

type Hasher interface {
	Hash(pw string) (string, error)
	Compare(hash, plain string) error
}
