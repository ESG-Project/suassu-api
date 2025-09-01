package feature

import (
	"context"
)

type Repo interface {
	Upsert(ctx context.Context, name string) error
}

type Hasher interface {
	Hash(pw string) (string, error)
	Compare(hash, plain string) error
}
