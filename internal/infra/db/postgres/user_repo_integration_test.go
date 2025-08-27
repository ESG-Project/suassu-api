//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	pg "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestUserRepo_CreateAndGet(t *testing.T) {
	ctx := context.Background()

	// 1) Postgres efêmero
	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// 2) Conexão
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// 3) Migration mínima (tabela "User" compatível com suas queries)
	const createUser = `
	CREATE TABLE "User" (
	  id text PRIMARY KEY,
	  name text NOT NULL,
	  email text NOT NULL,
	  password text NOT NULL,
	  document text NOT NULL,
	  phone text,
	  "addressId" text,
	  "roleId" text,
	  "enterpriseId" text NOT NULL
	);`
	_, err = db.Exec(createUser)
	require.NoError(t, err)

	// 4) Exercitar o repo real
	repo := pg.NewUserRepo(db)
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	u := domainuser.NewUser("u1", "Ana", "ana@ex.com", "HASH", "123", "e1")
	require.NoError(t, repo.Create(cctx, u))

	got, err := repo.GetByEmailForAuth(cctx, "ana@ex.com")
	require.NoError(t, err)
	require.Equal(t, "Ana", got.Name)
	require.Equal(t, "HASH", got.PasswordHash)
}
