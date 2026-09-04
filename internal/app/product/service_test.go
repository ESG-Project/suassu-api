package product_test

import (
	"context"
	"testing"

	appproduct "github.com/ESG-Project/suassu-api/internal/app/product"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainproduct "github.com/ESG-Project/suassu-api/internal/domain/product"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	byID      map[string]domainproduct.Product
	rows      []types.ProductDetailRow
	created   *domainproduct.Product
	updated   *domainproduct.Product
	deletedID string
}

func newFakeRepo(existing ...domainproduct.Product) *fakeRepo {
	m := make(map[string]domainproduct.Product, len(existing))
	for _, p := range existing {
		m[p.ID] = p
	}
	return &fakeRepo{byID: m}
}

func (f *fakeRepo) Create(ctx context.Context, p *domainproduct.Product) error {
	f.created = p
	f.byID[p.ID] = *p
	return nil
}

func (f *fakeRepo) GetByIDAnyEnterprise(ctx context.Context, id string) (*domainproduct.Product, error) {
	v, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	return &v, nil
}

func (f *fakeRepo) ListDetailedByEnterprise(ctx context.Context, enterpriseID string) ([]types.ProductDetailRow, error) {
	return f.rows, nil
}

func (f *fakeRepo) Update(ctx context.Context, p *domainproduct.Product) error {
	f.updated = p
	f.byID[p.ID] = *p
	return nil
}

func (f *fakeRepo) Delete(ctx context.Context, id, enterpriseID string) error {
	f.deletedID = id
	delete(f.byID, id)
	return nil
}

type fakeAudit struct{ entries []string }

func (f *fakeAudit) Record(ctx context.Context, actorID *string, enterpriseID, description string) {
	f.entries = append(f.entries, description)
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func existingProduct() domainproduct.Product {
	return domainproduct.Product{
		ID: "prod-1", Name: "Antigo", SuggestedValue: strPtr("100"), EnterpriseID: "ent-1",
		Deliverable: true, TypeProductID: strPtr("tp-1"), IsDefault: true,
	}
}

func TestNormalizeSuggestedValue(t *testing.T) {
	t.Parallel()

	t.Run("preserva a string original quando é um número válido", func(t *testing.T) {
		got, err := appproduct.NormalizeSuggestedValue(strPtr("1.234,56"))
		require.NoError(t, err)
		require.Equal(t, "1.234,56", *got, "o banco guarda o formato pt-BR como veio")
	})

	t.Run("vazio, ausente e o literal undefined viram nulo", func(t *testing.T) {
		for _, in := range []*string{nil, strPtr(""), strPtr("   "), strPtr("undefined")} {
			got, err := appproduct.NormalizeSuggestedValue(in)
			require.NoError(t, err)
			require.Nil(t, got)
		}
	})

	t.Run("recusa valor não numérico", func(t *testing.T) {
		_, err := appproduct.NormalizeSuggestedValue(strPtr("abc"))
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})

	t.Run("recusa valor negativo", func(t *testing.T) {
		_, err := appproduct.NormalizeSuggestedValue(strPtr("-1"))
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
		require.Contains(t, err.Error(), "negativo")
	})
}

func TestService_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("cria e registra na auditoria", func(t *testing.T) {
		repo, audit := newFakeRepo(), &fakeAudit{}
		svc := appproduct.NewService(repo, audit)

		p, err := svc.Create(ctx, "user-1", "ent-1", appproduct.CreateInput{
			Name: "Licenciamento", SuggestedValue: strPtr("2500"), TypeProductID: strPtr("tp-1"), Deliverable: true,
		})
		require.NoError(t, err)
		require.NotEmpty(t, p.ID)
		require.Equal(t, "2500", *p.SuggestedValue)
		require.Equal(t, "tp-1", *p.TypeProductID)
		require.False(t, p.IsDefault, "isDefault ausente cria produto não-padrão")
		require.Equal(t, []string{"Registro de produto Licenciamento, "}, audit.entries)
	})

	t.Run("tipo vazio vira nulo em vez de violar a FK", func(t *testing.T) {
		repo := newFakeRepo()
		svc := appproduct.NewService(repo, &fakeAudit{})

		p, err := svc.Create(ctx, "user-1", "ent-1", appproduct.CreateInput{Name: "X", TypeProductID: strPtr("")})
		require.NoError(t, err)
		require.Nil(t, p.TypeProductID)
	})

	t.Run("exige nome", func(t *testing.T) {
		svc := appproduct.NewService(newFakeRepo(), &fakeAudit{})

		_, err := svc.Create(ctx, "user-1", "ent-1", appproduct.CreateInput{Name: " "})
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})

	t.Run("valor inválido barra antes de gravar", func(t *testing.T) {
		repo := newFakeRepo()
		svc := appproduct.NewService(repo, &fakeAudit{})

		_, err := svc.Create(ctx, "user-1", "ent-1", appproduct.CreateInput{Name: "X", SuggestedValue: strPtr("-5")})
		require.Error(t, err)
		require.Nil(t, repo.created)
	})
}

func TestService_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("atualiza os campos enviados", func(t *testing.T) {
		repo, audit := newFakeRepo(existingProduct()), &fakeAudit{}
		svc := appproduct.NewService(repo, audit)

		p, err := svc.Update(ctx, "user-1", "ent-1", appproduct.UpdateInput{
			ID: "prod-1", Name: "Novo", SuggestedValue: strPtr("300"), Deliverable: boolPtr(false),
		})
		require.NoError(t, err)
		require.Equal(t, "Novo", p.Name)
		require.Equal(t, "300", *p.SuggestedValue)
		require.False(t, p.Deliverable)
		require.Equal(t, []string{"Edição do produto Novo"}, audit.entries)
	})

	t.Run("campos ausentes preservam o valor atual", func(t *testing.T) {
		repo := newFakeRepo(existingProduct())
		svc := appproduct.NewService(repo, &fakeAudit{})

		p, err := svc.Update(ctx, "user-1", "ent-1", appproduct.UpdateInput{ID: "prod-1", Name: "Novo"})
		require.NoError(t, err)
		require.True(t, p.Deliverable)
		require.True(t, p.IsDefault)
		require.Equal(t, "tp-1", *p.TypeProductID)
	})

	t.Run("tipo enviado vazio limpa a associação", func(t *testing.T) {
		repo := newFakeRepo(existingProduct())
		svc := appproduct.NewService(repo, &fakeAudit{})

		p, err := svc.Update(ctx, "user-1", "ent-1", appproduct.UpdateInput{
			ID: "prod-1", Name: "Novo", TypeProductID: strPtr(""),
		})
		require.NoError(t, err)
		require.Nil(t, p.TypeProductID)
	})

	t.Run("404 quando não existe", func(t *testing.T) {
		svc := appproduct.NewService(newFakeRepo(), &fakeAudit{})

		_, err := svc.Update(ctx, "user-1", "ent-1", appproduct.UpdateInput{ID: "sumiu", Name: "Novo"})
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})

	t.Run("403 quando pertence a outra empresa", func(t *testing.T) {
		repo := newFakeRepo(existingProduct())
		svc := appproduct.NewService(repo, &fakeAudit{})

		_, err := svc.Update(ctx, "user-1", "ent-2", appproduct.UpdateInput{ID: "prod-1", Name: "Novo"})
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.Nil(t, repo.updated)
	})
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("apaga o produto da própria empresa", func(t *testing.T) {
		repo, audit := newFakeRepo(existingProduct()), &fakeAudit{}
		svc := appproduct.NewService(repo, audit)

		require.NoError(t, svc.Delete(ctx, "user-1", "ent-1", "prod-1"))
		require.Equal(t, "prod-1", repo.deletedID)
		require.Equal(t, []string{"Exclusão de produto Antigo"}, audit.entries)
	})

	t.Run("403 quando pertence a outra empresa", func(t *testing.T) {
		repo := newFakeRepo(existingProduct())
		svc := appproduct.NewService(repo, &fakeAudit{})

		err := svc.Delete(ctx, "user-1", "ent-2", "prod-1")
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.Empty(t, repo.deletedID)
	})
}

func TestService_List(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	repo.rows = []types.ProductDetailRow{{
		ID: "prod-1", Name: "X", EnterpriseID: "ent-1",
		Parameter: &types.ProductParameterRef{ID: "par-1", Title: "Área", Value: strPtr("10")},
		Type:      &types.ProductTypeRef{ID: "tp-1", Type: "Consultoria"},
	}}
	svc := appproduct.NewService(repo, &fakeAudit{})

	rows, err := svc.List(context.Background(), "ent-1")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "Área", rows[0].Parameter.Title)
	require.Equal(t, "Consultoria", rows[0].Type.Type)
}
