package postgres

import (
	"context"
	"database/sql"

	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

// Adaptador para permitir *sql.DB e *sql.Tx via sqlc.DBTX
type dbtx = sqlc.DBTX

type TxManager struct{ DB *sql.DB }

type Repos struct {
	Users     func() *UserRepo
	Addresses func() *AddressRepo
}

func (m *TxManager) RunInTx(ctx context.Context, fn func(r Repos) error) error {
	tx, err := m.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	r := Repos{
		Users:     func() *UserRepo { return NewUserRepoFrom(tx) },
		Addresses: func() *AddressRepo { return NewAddressRepoFrom(tx) },
	}

	if err := fn(r); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
