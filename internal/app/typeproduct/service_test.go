package typeproduct_test

import (
	"context"
	"testing"

	apptypeproduct "github.com/ESG-Project/suassu-api/internal/app/typeproduct"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domaintypeproduct "github.com/ESG-Project/suassu-api/internal/domain/typeproduct"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	byID      map[string]domaintypeproduct.TypeProduct
	created   *domaintypeproduct.TypeProduct
	updated   *domaintypeproduct.TypeProduct
	deletedID string
}

func newFakeRepo(existing ...domaintypeproduct.TypeProduct) *fakeRepo {
	m := make(map[string]domaintypeproduct.TypeProduct, len(existing))
	for _, t := range existing {
		m[t.ID] = t
	}
	return &fakeRepo{byID: m}
}

func (f *fakeRepo) Create(ctx context.Context, t *domaintypeproduct.TypeProduct) error {
	f.created = t
	f.byID[t.ID] = *t
	return nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id string) (*domaintypeproduct.TypeProduct, error) {
	v, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	return &v, nil
}

func (f *fakeRepo) List(ctx context.Context, enterpriseID string) ([]domaintypeproduct.TypeProduct, error) {
	out := []domaintypeproduct.TypeProduct{}
	for _, v := range f.byID {
		if v.EnterpriseID == enterpriseID {
			out = append(out, v)
		}
	}
	return out, nil
}

func (f *fakeRepo) Update(ctx context.Context, t *domaintypeproduct.TypeProduct) error {
	f.updated = t
	f.byID[t.ID] = *t
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

func TestService_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("cria e registra na auditoria", func(t *testing.T) {
		repo, audit := newFakeRepo(), &fakeAudit{}
		svc := apptypeproduct.NewService(repo, audit)

		tp, err := svc.Create(ctx, "user-1", "ent-1", "Consultoria")
		require.NoError(t, err)
		require.NotEmpty(t, tp.ID)
		require.Equal(t, "ent-1", tp.EnterpriseID)
		require.Equal(t, []string{"Registro de um tipo de produto Consultoria"}, audit.entries)
	})

	t.Run("exige o tipo", func(t *testing.T) {
		svc := apptypeproduct.NewService(newFakeRepo(), &fakeAudit{})

		_, err := svc.Create(ctx, "user-1", "ent-1", "   ")
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})
}

func TestService_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("atualiza o tipo da própria empresa", func(t *testing.T) {
		repo := newFakeRepo(domaintypeproduct.TypeProduct{ID: "tp-1", Type: "Antigo", EnterpriseID: "ent-1"})
		audit := &fakeAudit{}
		svc := apptypeproduct.NewService(repo, audit)

		tp, err := svc.Update(ctx, "user-1", "ent-1", "tp-1", "Novo")
		require.NoError(t, err)
		require.Equal(t, "Novo", tp.Type)
		require.Equal(t, "Novo", repo.updated.Type)
		require.Equal(t, []string{"Edição de um tipo de produto Novo"}, audit.entries)
	})

	t.Run("404 quando não existe", func(t *testing.T) {
		svc := apptypeproduct.NewService(newFakeRepo(), &fakeAudit{})

		_, err := svc.Update(ctx, "user-1", "ent-1", "sumiu", "Novo")
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})

	t.Run("403 quando pertence a outra empresa", func(t *testing.T) {
		repo := newFakeRepo(domaintypeproduct.TypeProduct{ID: "tp-1", Type: "Antigo", EnterpriseID: "ent-2"})
		svc := apptypeproduct.NewService(repo, &fakeAudit{})

		_, err := svc.Update(ctx, "user-1", "ent-1", "tp-1", "Novo")
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.Nil(t, repo.updated)
	})
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("apaga o tipo da própria empresa", func(t *testing.T) {
		repo := newFakeRepo(domaintypeproduct.TypeProduct{ID: "tp-1", Type: "Consultoria", EnterpriseID: "ent-1"})
		audit := &fakeAudit{}
		svc := apptypeproduct.NewService(repo, audit)

		require.NoError(t, svc.Delete(ctx, "user-1", "ent-1", "tp-1"))
		require.Equal(t, "tp-1", repo.deletedID)
		require.Equal(t, []string{"Exclusão de um tipo de produto Consultoria"}, audit.entries)
	})

	t.Run("403 quando pertence a outra empresa", func(t *testing.T) {
		repo := newFakeRepo(domaintypeproduct.TypeProduct{ID: "tp-1", Type: "X", EnterpriseID: "ent-2"})
		svc := apptypeproduct.NewService(repo, &fakeAudit{})

		err := svc.Delete(ctx, "user-1", "ent-1", "tp-1")
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.Empty(t, repo.deletedID)
	})
}
