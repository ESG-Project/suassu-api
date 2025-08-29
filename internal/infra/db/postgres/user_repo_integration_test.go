//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
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

func Test_RunInTx_CreateUserAndAddress_Atomicity(t *testing.T) {
	ctx := context.Background()

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

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// schema mínimo para Address e User
	const schema = `
	CREATE TABLE "Address" (
	  id text PRIMARY KEY,
	  "zipCode" text NOT NULL,
	  state text NOT NULL,
	  city text NOT NULL,
	  neighborhood text NOT NULL,
	  street text NOT NULL,
	  num text NOT NULL,
	  latitude text,
	  longitude text,
	  "addInfo" text
	);
	CREATE TABLE "User" (
	  id text PRIMARY KEY,
	  name text NOT NULL,
	  email text NOT NULL,
	  password text NOT NULL,
	  document text NOT NULL,
	  phone text,
	  "addressId" text,
	  "roleId" text,
	  "enterpriseId" text NOT NULL,
	  FOREIGN KEY ("addressId") REFERENCES "Address"(id)
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	txm := &pg.TxManager{DB: db}

	// sucesso: cria address e user dentro da mesma tx
	err = txm.RunInTx(ctx, func(r pg.Repos) error {
		addr := &domainaddress.Address{ID: "a1", ZipCode: "00000-000", State: "SP", City: "SP", Neighborhood: "C", Street: "S", Num: "1"}
		if err := r.Addresses().Create(ctx, addr); err != nil {
			return err
		}
		u := domainuser.NewUser("u2", "Bob", "bob@ex.com", "H", "D", "e1")
		u.SetAddressID(&addr.ID)
		return r.Users().Create(ctx, u)
	})
	require.NoError(t, err)

	// rollback: falha ao criar usuário força rollback do address
	err = txm.RunInTx(ctx, func(r pg.Repos) error {
		addr := &domainaddress.Address{ID: "a2", ZipCode: "00000-000", State: "SP", City: "SP", Neighborhood: "C", Street: "S", Num: "1"}
		if err := r.Addresses().Create(ctx, addr); err != nil {
			return err
		}
		return errors.New("force fail")
	})
	require.Error(t, err)

	// Verifica se a2 não foi persistido
	var count int
	err = db.QueryRowContext(ctx, `SELECT COUNT(1) FROM "Address" WHERE id = 'a2'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}
