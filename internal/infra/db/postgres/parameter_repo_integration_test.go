//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	domainparameter "github.com/ESG-Project/suassu-api/internal/domain/parameter"
	pg "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
)

func TestParameterRepo_CreateAndGet(t *testing.T) {
	ctx := context.Background()

	// Setup container
	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	time.Sleep(2 * time.Second)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	err = db.PingContext(ctx)
	require.NoError(t, err)

	// Schema
	const schema = `
	CREATE TABLE "Enterprise" (
		id text PRIMARY KEY,
		cnpj text NOT NULL,
		"addressId" text,
		email text NOT NULL,
		"fantasyName" text,
		name text NOT NULL,
		phone text
	);
	CREATE TABLE "Parameter" (
		id text PRIMARY KEY,
		title text NOT NULL,
		value text,
		"enterpriseId" text NOT NULL,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	// Criar enterprise de teste
	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	// Testar repositório
	repo := pg.NewParameterRepo(db)

	// Create
	value := "Test Value"
	param := domainparameter.NewParameter(uuid.NewString(), "Test Parameter", enterpriseID)
	param.SetValue(&value)
	err = repo.Create(ctx, param)
	require.NoError(t, err)

	// Get
	fetched, err := repo.GetByID(ctx, param.ID, enterpriseID)
	require.NoError(t, err)
	require.Equal(t, param.Title, fetched.Title)
	require.Equal(t, *param.Value, *fetched.Value)
	require.Equal(t, param.EnterpriseID, fetched.EnterpriseID)
	require.False(t, fetched.IsDefault)
}

func TestParameterRepo_List(t *testing.T) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	time.Sleep(2 * time.Second)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	const schema = `
	CREATE TABLE "Enterprise" (
		id text PRIMARY KEY,
		cnpj text NOT NULL,
		"addressId" text,
		email text NOT NULL,
		"fantasyName" text,
		name text NOT NULL,
		phone text
	);
	CREATE TABLE "Parameter" (
		id text PRIMARY KEY,
		title text NOT NULL,
		value text,
		"enterpriseId" text NOT NULL,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	repo := pg.NewParameterRepo(db)

	// Criar múltiplos parâmetros
	param1 := domainparameter.NewParameter(uuid.NewString(), "Param A", enterpriseID)
	err = repo.Create(ctx, param1)
	require.NoError(t, err)

	param2 := domainparameter.NewParameter(uuid.NewString(), "Param B", enterpriseID)
	param2.SetIsDefault(true)
	err = repo.Create(ctx, param2)
	require.NoError(t, err)

	// Listar
	list, err := repo.List(ctx, enterpriseID)
	require.NoError(t, err)
	require.Len(t, list, 2)
}

func TestParameterRepo_Update(t *testing.T) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	time.Sleep(2 * time.Second)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	const schema = `
	CREATE TABLE "Enterprise" (
		id text PRIMARY KEY,
		cnpj text NOT NULL,
		"addressId" text,
		email text NOT NULL,
		"fantasyName" text,
		name text NOT NULL,
		phone text
	);
	CREATE TABLE "Parameter" (
		id text PRIMARY KEY,
		title text NOT NULL,
		value text,
		"enterpriseId" text NOT NULL,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	repo := pg.NewParameterRepo(db)

	// Criar
	value := "Original Value"
	param := domainparameter.NewParameter(uuid.NewString(), "Original Title", enterpriseID)
	param.SetValue(&value)
	err = repo.Create(ctx, param)
	require.NoError(t, err)

	// Atualizar
	newValue := "Updated Value"
	param.Title = "Updated Title"
	param.SetValue(&newValue)
	param.SetIsDefault(true)

	err = repo.Update(ctx, param)
	require.NoError(t, err)

	// Verificar
	fetched, err := repo.GetByID(ctx, param.ID, enterpriseID)
	require.NoError(t, err)
	require.Equal(t, "Updated Title", fetched.Title)
	require.Equal(t, "Updated Value", *fetched.Value)
	require.True(t, fetched.IsDefault)
}

func TestParameterRepo_Delete(t *testing.T) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	time.Sleep(2 * time.Second)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	const schema = `
	CREATE TABLE "Enterprise" (
		id text PRIMARY KEY,
		cnpj text NOT NULL,
		"addressId" text,
		email text NOT NULL,
		"fantasyName" text,
		name text NOT NULL,
		phone text
	);
	CREATE TABLE "Parameter" (
		id text PRIMARY KEY,
		title text NOT NULL,
		value text,
		"enterpriseId" text NOT NULL,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	repo := pg.NewParameterRepo(db)

	// Criar
	param := domainparameter.NewParameter(uuid.NewString(), "To Delete", enterpriseID)
	err = repo.Create(ctx, param)
	require.NoError(t, err)

	// Deletar
	err = repo.Delete(ctx, param.ID, enterpriseID)
	require.NoError(t, err)

	// Verificar que não existe mais
	_, err = repo.GetByID(ctx, param.ID, enterpriseID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestParameterRepo_MultiTenant_Isolation(t *testing.T) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	time.Sleep(2 * time.Second)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	const schema = `
	CREATE TABLE "Enterprise" (
		id text PRIMARY KEY,
		cnpj text NOT NULL,
		"addressId" text,
		email text NOT NULL,
		"fantasyName" text,
		name text NOT NULL,
		phone text
	);
	CREATE TABLE "Parameter" (
		id text PRIMARY KEY,
		title text NOT NULL,
		value text,
		"enterpriseId" text NOT NULL,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	// Criar dois enterprises
	enterprise1ID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterprise1ID, "11111111000100", "test1@test.com", "Enterprise 1")
	require.NoError(t, err)

	enterprise2ID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterprise2ID, "22222222000100", "test2@test.com", "Enterprise 2")
	require.NoError(t, err)

	repo := pg.NewParameterRepo(db)

	// Criar parâmetro para enterprise1
	param := domainparameter.NewParameter(uuid.NewString(), "Enterprise 1 Param", enterprise1ID)
	err = repo.Create(ctx, param)
	require.NoError(t, err)

	// Tentar acessar com enterprise2 (deve falhar)
	_, err = repo.GetByID(ctx, param.ID, enterprise2ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")

	// Listar para enterprise2 (não deve incluir param de enterprise1)
	list, err := repo.List(ctx, enterprise2ID)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestParameterRepo_CascadeDelete(t *testing.T) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	time.Sleep(2 * time.Second)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	const schema = `
	CREATE TABLE "Enterprise" (
		id text PRIMARY KEY,
		cnpj text NOT NULL,
		"addressId" text,
		email text NOT NULL,
		"fantasyName" text,
		name text NOT NULL,
		phone text
	);
	CREATE TABLE "Parameter" (
		id text PRIMARY KEY,
		title text NOT NULL,
		value text,
		"enterpriseId" text NOT NULL,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	repo := pg.NewParameterRepo(db)

	// Criar parâmetro
	param := domainparameter.NewParameter(uuid.NewString(), "Cascade Test", enterpriseID)
	err = repo.Create(ctx, param)
	require.NoError(t, err)

	// Deletar enterprise (deve deletar parâmetro também por CASCADE)
	_, err = db.ExecContext(ctx, `DELETE FROM "Enterprise" WHERE id = $1`, enterpriseID)
	require.NoError(t, err)

	// Verificar que parâmetro foi deletado
	_, err = repo.GetByID(ctx, param.ID, enterpriseID)
	require.Error(t, err)
}
