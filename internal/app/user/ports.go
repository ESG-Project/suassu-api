package user

import "context"

type Entity struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Document     string
	Phone        *string
	AddressID    *string
	RoleID       *string
	EnterpriseID string
}

type Repo interface {
	Create(ctx context.Context, u Entity) error
	GetByEmail(ctx context.Context, email string) (Entity, error)
	List(ctx context.Context, limit, offset int32) ([]Entity, error)
}

type Hasher interface {
	Hash(pw string) (string, error)
}
