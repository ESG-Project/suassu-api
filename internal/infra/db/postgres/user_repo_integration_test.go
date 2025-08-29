//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/address"
	"github.com/ESG-Project/suassu-api/internal/app/user"
	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	infraauth "github.com/ESG-Project/suassu-api/internal/infra/auth"
	pg "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestUserRepo_CreateAndGet(t *testing.T) {
	ctx := context.Background()

	// 1) Postgres efêmero - Testcontainers escolhe porta automaticamente
	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	// Aguarda um pouco para o container estar totalmente pronto
	time.Sleep(2 * time.Second)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// 2) Conexão
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// Ping para verificar conexão
	err = db.PingContext(ctx)
	require.NoError(t, err)

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

	// Aguarda um pouco para o container estar totalmente pronto
	time.Sleep(2 * time.Second)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// Ping para verificar conexão
	err = db.PingContext(ctx)
	require.NoError(t, err)

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

func Test_ServiceWithTxManager_CreateUserWithAddress(t *testing.T) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	// Aguarda um pouco para o container estar totalmente pronto
	time.Sleep(2 * time.Second)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// Ping para verificar conexão
	err = db.PingContext(ctx)
	require.NoError(t, err)

	// schema completo para Address e User
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

	// Setup services com TxManager
	txm := &pg.TxManager{DB: db}
	userRepo := pg.NewUserRepo(db)
	addressRepo := pg.NewAddressRepo(db)
	hasher := infraauth.NewBCrypt()
	addressSvc := address.NewService(addressRepo, hasher)
	userSvc := user.NewServiceWithTx(userRepo, addressSvc, hasher, txm)

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	t.Run("create user with address using transaction", func(t *testing.T) {
		// Cria usuário com endereço (deve usar transação)
		id, err := userSvc.Create(cctx, "ent-1", user.CreateInput{
			Name: "João", Email: "joao@ex.com", Password: "123",
			Document: "999", EnterpriseID: "ent-1",
			Address: &address.CreateInput{
				ZipCode: "00000-000", State: "SP", City: "SP",
				Neighborhood: "Centro", Street: "Rua A", Num: "123",
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// Verifica se ambos foram criados
		var addrCount, userCount int
		err = db.QueryRowContext(cctx, `SELECT COUNT(1) FROM "Address" WHERE "zipCode" = '00000-000'`).Scan(&addrCount)
		require.NoError(t, err)
		err = db.QueryRowContext(cctx, `SELECT COUNT(1) FROM "User" WHERE id = $1`, id).Scan(&userCount)
		require.NoError(t, err)
		require.Equal(t, 1, addrCount)
		require.Equal(t, 1, userCount)
	})

	t.Run("create user with existing addressId (no transaction)", func(t *testing.T) {
		// Primeiro cria um endereço
		addr := &domainaddress.Address{ID: "existing-addr", ZipCode: "11111-111", State: "RJ", City: "RJ", Neighborhood: "C", Street: "S", Num: "1"}
		err := addressRepo.Create(cctx, addr)
		require.NoError(t, err)

		// Cria usuário referenciando endereço existente
		id, err := userSvc.Create(cctx, "ent-1", user.CreateInput{
			Name: "Maria", Email: "maria@ex.com", Password: "123",
			Document: "888", EnterpriseID: "ent-1",
			AddressID: &addr.ID,
		})
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// Verifica se usuário foi criado
		var userCount int
		err = db.QueryRowContext(cctx, `SELECT COUNT(1) FROM "User" WHERE id = $1`, id).Scan(&userCount)
		require.NoError(t, err)
		require.Equal(t, 1, userCount)
	})
}

func Test_TxManager_RollbackOnError(t *testing.T) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	// Aguarda um pouco para o container estar totalmente pronto
	time.Sleep(2 * time.Second)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// Ping para verificar conexão
	err = db.PingContext(ctx)
	require.NoError(t, err)

	// schema mínimo
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
	  "enterpriseId" text NOT NULL,
	  "addressId" text,
	  "roleId" text,
	  FOREIGN KEY ("addressId") REFERENCES "Address"(id)
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	txm := &pg.TxManager{DB: db}
	userRepo := pg.NewUserRepo(db)

	t.Run("rollback when user creation fails", func(t *testing.T) {
		// Primeiro cria um usuário válido
		validUser := domainuser.NewUser("rollback-user", "Valid User", "valid@ex.com", "HASH", "123", "ent-1")
		err := userRepo.Create(ctx, validUser)
		require.NoError(t, err)

		// Tenta criar endereço + usuário, mas usuário falha (ID duplicado)
		err = txm.RunInTx(ctx, func(r pg.Repos) error {
			// Cria endereço com sucesso
			addr := &domainaddress.Address{ID: "rollback-addr", ZipCode: "22222-222", State: "MG", City: "BH", Neighborhood: "C", Street: "S", Num: "1"}
			if err := r.Addresses().Create(ctx, addr); err != nil {
				return err
			}

			// Tenta criar usuário com ID duplicado (deve falhar por constraint)
			u := domainuser.NewUser("rollback-user", "Duplicate User", "duplicate@ex.com", "HASH", "456", "ent-1")
			return r.Users().Create(ctx, u)
		})
		require.Error(t, err)

		// Verifica se NADA foi persistido (rollback funcionou)
		var addrCount int
		err = db.QueryRowContext(ctx, `SELECT COUNT(1) FROM "Address" WHERE id = 'rollback-addr'`).Scan(&addrCount)
		require.NoError(t, err)
		require.Equal(t, 0, addrCount)
	})

	t.Run("commit when everything succeeds", func(t *testing.T) {
		// Cria endereço + usuário com sucesso
		err := txm.RunInTx(ctx, func(r pg.Repos) error {
			addr := &domainaddress.Address{ID: "commit-addr", ZipCode: "33333-333", State: "RS", City: "POA", Neighborhood: "C", Street: "S", Num: "1"}
			if err := r.Addresses().Create(ctx, addr); err != nil {
				return err
			}

			u := domainuser.NewUser("commit-user", "Commit User", "commit@ex.com", "HASH", "123", "ent-1")
			u.SetAddressID(&addr.ID)
			return r.Users().Create(ctx, u)
		})
		require.NoError(t, err)

		// Verifica se AMBOS foram persistidos (commit funcionou)
		var addrCount, userCount int
		err = db.QueryRowContext(ctx, `SELECT COUNT(1) FROM "Address" WHERE id = 'commit-addr'`).Scan(&addrCount)
		require.NoError(t, err)
		err = db.QueryRowContext(ctx, `SELECT COUNT(1) FROM "User" WHERE id = 'commit-user'`).Scan(&userCount)
		require.NoError(t, err)
		require.Equal(t, 1, addrCount)
		require.Equal(t, 1, userCount)
	})
}
