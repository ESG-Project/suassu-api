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

	domainproduct "github.com/ESG-Project/suassu-api/internal/domain/product"
	pg "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
)

func TestProductRepo_CreateAndGet(t *testing.T) {
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
	);
	CREATE TABLE "Product" (
		id text PRIMARY KEY,
		name text NOT NULL,
		"suggestedValue" text,
		"enterpriseId" text NOT NULL,
		"parameterId" text,
		deliverable boolean NOT NULL,
		"typeProductId" text,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE,
		FOREIGN KEY ("parameterId") REFERENCES "Parameter"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	// Criar enterprise de teste
	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	// Testar repositório
	repo := pg.NewProductRepo(db)

	// Create
	suggestedValue := "100.00"
	product := domainproduct.NewProduct(uuid.NewString(), "Test Product", enterpriseID, true)
	product.SetSuggestedValue(&suggestedValue)
	err = repo.Create(ctx, product)
	require.NoError(t, err)

	// Get
	fetched, err := repo.GetByID(ctx, product.ID, enterpriseID)
	require.NoError(t, err)
	require.Equal(t, product.Name, fetched.Name)
	require.Equal(t, *product.SuggestedValue, *fetched.SuggestedValue)
	require.Equal(t, product.EnterpriseID, fetched.EnterpriseID)
	require.True(t, fetched.Deliverable)
	require.False(t, fetched.IsDefault)
}

func TestProductRepo_List(t *testing.T) {
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
	);
	CREATE TABLE "Product" (
		id text PRIMARY KEY,
		name text NOT NULL,
		"suggestedValue" text,
		"enterpriseId" text NOT NULL,
		"parameterId" text,
		deliverable boolean NOT NULL,
		"typeProductId" text,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE,
		FOREIGN KEY ("parameterId") REFERENCES "Parameter"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	repo := pg.NewProductRepo(db)

	// Criar múltiplos produtos
	product1 := domainproduct.NewProduct(uuid.NewString(), "Product A", enterpriseID, true)
	err = repo.Create(ctx, product1)
	require.NoError(t, err)

	product2 := domainproduct.NewProduct(uuid.NewString(), "Product B", enterpriseID, false)
	product2.SetIsDefault(true)
	err = repo.Create(ctx, product2)
	require.NoError(t, err)

	// Listar
	list, err := repo.List(ctx, enterpriseID)
	require.NoError(t, err)
	require.Len(t, list, 2)
}

func TestProductRepo_Update(t *testing.T) {
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
	);
	CREATE TABLE "Product" (
		id text PRIMARY KEY,
		name text NOT NULL,
		"suggestedValue" text,
		"enterpriseId" text NOT NULL,
		"parameterId" text,
		deliverable boolean NOT NULL,
		"typeProductId" text,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE,
		FOREIGN KEY ("parameterId") REFERENCES "Parameter"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	repo := pg.NewProductRepo(db)

	// Criar
	suggestedValue := "50.00"
	product := domainproduct.NewProduct(uuid.NewString(), "Original Product", enterpriseID, false)
	product.SetSuggestedValue(&suggestedValue)
	err = repo.Create(ctx, product)
	require.NoError(t, err)

	// Atualizar
	newValue := "150.00"
	product.Name = "Updated Product"
	product.SetSuggestedValue(&newValue)
	product.Deliverable = true
	product.SetIsDefault(true)

	err = repo.Update(ctx, product)
	require.NoError(t, err)

	// Verificar
	fetched, err := repo.GetByID(ctx, product.ID, enterpriseID)
	require.NoError(t, err)
	require.Equal(t, "Updated Product", fetched.Name)
	require.Equal(t, "150.00", *fetched.SuggestedValue)
	require.True(t, fetched.Deliverable)
	require.True(t, fetched.IsDefault)
}

func TestProductRepo_Delete(t *testing.T) {
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
	);
	CREATE TABLE "Product" (
		id text PRIMARY KEY,
		name text NOT NULL,
		"suggestedValue" text,
		"enterpriseId" text NOT NULL,
		"parameterId" text,
		deliverable boolean NOT NULL,
		"typeProductId" text,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE,
		FOREIGN KEY ("parameterId") REFERENCES "Parameter"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	repo := pg.NewProductRepo(db)

	// Criar
	product := domainproduct.NewProduct(uuid.NewString(), "To Delete", enterpriseID, true)
	err = repo.Create(ctx, product)
	require.NoError(t, err)

	// Deletar
	err = repo.Delete(ctx, product.ID, enterpriseID)
	require.NoError(t, err)

	// Verificar que não existe mais
	_, err = repo.GetByID(ctx, product.ID, enterpriseID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestProductRepo_MultiTenant_Isolation(t *testing.T) {
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
	);
	CREATE TABLE "Product" (
		id text PRIMARY KEY,
		name text NOT NULL,
		"suggestedValue" text,
		"enterpriseId" text NOT NULL,
		"parameterId" text,
		deliverable boolean NOT NULL,
		"typeProductId" text,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE,
		FOREIGN KEY ("parameterId") REFERENCES "Parameter"(id) ON DELETE CASCADE
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

	repo := pg.NewProductRepo(db)

	// Criar produto para enterprise1
	product := domainproduct.NewProduct(uuid.NewString(), "Enterprise 1 Product", enterprise1ID, true)
	err = repo.Create(ctx, product)
	require.NoError(t, err)

	// Tentar acessar com enterprise2 (deve falhar)
	_, err = repo.GetByID(ctx, product.ID, enterprise2ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")

	// Listar para enterprise2 (não deve incluir produto de enterprise1)
	list, err := repo.List(ctx, enterprise2ID)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestProductRepo_CascadeDelete_FromEnterprise(t *testing.T) {
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
	);
	CREATE TABLE "Product" (
		id text PRIMARY KEY,
		name text NOT NULL,
		"suggestedValue" text,
		"enterpriseId" text NOT NULL,
		"parameterId" text,
		deliverable boolean NOT NULL,
		"typeProductId" text,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE,
		FOREIGN KEY ("parameterId") REFERENCES "Parameter"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	repo := pg.NewProductRepo(db)

	// Criar produto
	product := domainproduct.NewProduct(uuid.NewString(), "Cascade Test", enterpriseID, true)
	err = repo.Create(ctx, product)
	require.NoError(t, err)

	// Deletar enterprise (deve deletar produto também por CASCADE)
	_, err = db.ExecContext(ctx, `DELETE FROM "Enterprise" WHERE id = $1`, enterpriseID)
	require.NoError(t, err)

	// Verificar que produto foi deletado
	_, err = repo.GetByID(ctx, product.ID, enterpriseID)
	require.Error(t, err)
}

func TestProductRepo_CascadeDelete_FromParameter(t *testing.T) {
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
	);
	CREATE TABLE "Product" (
		id text PRIMARY KEY,
		name text NOT NULL,
		"suggestedValue" text,
		"enterpriseId" text NOT NULL,
		"parameterId" text,
		deliverable boolean NOT NULL,
		"typeProductId" text,
		"isDefault" boolean NOT NULL DEFAULT false,
		FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise"(id) ON DELETE CASCADE,
		FOREIGN KEY ("parameterId") REFERENCES "Parameter"(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	enterpriseID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Enterprise" (id, cnpj, email, name) VALUES ($1, $2, $3, $4)`,
		enterpriseID, "12345678000100", "test@test.com", "Test Enterprise")
	require.NoError(t, err)

	// Criar parâmetro
	parameterID := uuid.NewString()
	_, err = db.Exec(`INSERT INTO "Parameter" (id, title, "enterpriseId", "isDefault") VALUES ($1, $2, $3, $4)`,
		parameterID, "Test Parameter", enterpriseID, false)
	require.NoError(t, err)

	repo := pg.NewProductRepo(db)

	// Criar produto com referência ao parâmetro
	product := domainproduct.NewProduct(uuid.NewString(), "Product With Parameter", enterpriseID, true)
	product.SetParameterID(&parameterID)
	err = repo.Create(ctx, product)
	require.NoError(t, err)

	// Deletar parâmetro (deve deletar produto também por CASCADE)
	_, err = db.ExecContext(ctx, `DELETE FROM "Parameter" WHERE id = $1`, parameterID)
	require.NoError(t, err)

	// Verificar que produto foi deletado
	_, err = repo.GetByID(ctx, product.ID, enterpriseID)
	require.Error(t, err)
}
