package parameter_test

import (
	"context"
	"testing"

	appparameter "github.com/ESG-Project/suassu-api/internal/app/parameter"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainparameter "github.com/ESG-Project/suassu-api/internal/domain/parameter"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	byID    map[string]domainparameter.Parameter
	updated *domainparameter.Parameter
}

func newFakeRepo(existing ...domainparameter.Parameter) *fakeRepo {
	m := make(map[string]domainparameter.Parameter, len(existing))
	for _, p := range existing {
		m[p.ID] = p
	}
	return &fakeRepo{byID: m}
}

func (f *fakeRepo) GetByIDAnyEnterprise(ctx context.Context, id string) (*domainparameter.Parameter, error) {
	v, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	return &v, nil
}

func (f *fakeRepo) List(ctx context.Context, enterpriseID string) ([]domainparameter.Parameter, error) {
	out := []domainparameter.Parameter{}
	for _, v := range f.byID {
		if v.EnterpriseID == enterpriseID {
			out = append(out, v)
		}
	}
	return out, nil
}

func (f *fakeRepo) Update(ctx context.Context, p *domainparameter.Parameter) error {
	f.updated = p
	f.byID[p.ID] = *p
	return nil
}

type fakeAudit struct{ entries []string }

func (f *fakeAudit) Record(ctx context.Context, actorID *string, enterpriseID, description string) {
	f.entries = append(f.entries, description)
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func existingParameter() domainparameter.Parameter {
	return domainparameter.Parameter{
		ID: "par-1", Title: "Antigo", Value: strPtr("10"), EnterpriseID: "ent-1", IsDefault: true,
	}
}

func TestService_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("atualiza título e valor", func(t *testing.T) {
		repo, audit := newFakeRepo(existingParameter()), &fakeAudit{}
		svc := appparameter.NewService(repo, audit)

		p, err := svc.Update(ctx, "user-1", "ent-1", appparameter.UpdateInput{
			ID: "par-1", Title: "Novo", Value: strPtr("25"),
		})
		require.NoError(t, err)
		require.Equal(t, "Novo", p.Title)
		require.Equal(t, "25", *p.Value)
		require.Equal(t, []string{"Edição do parâmetro Novo"}, audit.entries)
	})

	t.Run("isDefault ausente preserva o valor atual", func(t *testing.T) {
		repo := newFakeRepo(existingParameter())
		svc := appparameter.NewService(repo, &fakeAudit{})

		p, err := svc.Update(ctx, "user-1", "ent-1", appparameter.UpdateInput{ID: "par-1", Title: "Novo"})
		require.NoError(t, err)
		require.True(t, p.IsDefault, "o front não envia isDefault ao editar; não pode ser zerado")
	})

	t.Run("isDefault informado é aplicado", func(t *testing.T) {
		repo := newFakeRepo(existingParameter())
		svc := appparameter.NewService(repo, &fakeAudit{})

		p, err := svc.Update(ctx, "user-1", "ent-1", appparameter.UpdateInput{
			ID: "par-1", Title: "Novo", IsDefault: boolPtr(false),
		})
		require.NoError(t, err)
		require.False(t, p.IsDefault)
	})

	t.Run("value ausente preserva o valor atual", func(t *testing.T) {
		repo := newFakeRepo(existingParameter())
		svc := appparameter.NewService(repo, &fakeAudit{})

		p, err := svc.Update(ctx, "user-1", "ent-1", appparameter.UpdateInput{ID: "par-1", Title: "Novo"})
		require.NoError(t, err)
		require.Equal(t, "10", *p.Value)
	})

	t.Run("exige título", func(t *testing.T) {
		svc := appparameter.NewService(newFakeRepo(existingParameter()), &fakeAudit{})

		_, err := svc.Update(ctx, "user-1", "ent-1", appparameter.UpdateInput{ID: "par-1", Title: " "})
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})

	t.Run("404 quando não existe", func(t *testing.T) {
		svc := appparameter.NewService(newFakeRepo(), &fakeAudit{})

		_, err := svc.Update(ctx, "user-1", "ent-1", appparameter.UpdateInput{ID: "sumiu", Title: "Novo"})
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})

	t.Run("403 quando pertence a outra empresa", func(t *testing.T) {
		repo := newFakeRepo(existingParameter())
		svc := appparameter.NewService(repo, &fakeAudit{})

		_, err := svc.Update(ctx, "user-1", "ent-2", appparameter.UpdateInput{ID: "par-1", Title: "Novo"})
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.Nil(t, repo.updated)
	})
}
