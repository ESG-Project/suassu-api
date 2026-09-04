package feature_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ESG-Project/suassu-api/internal/app/feature"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainfeature "github.com/ESG-Project/suassu-api/internal/domain/feature"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	byID    map[string]domainfeature.Feature
	created *domainfeature.Feature
	deleted string
	err     error
}

func newFakeRepo(existing ...domainfeature.Feature) *fakeRepo {
	m := make(map[string]domainfeature.Feature, len(existing))
	for _, f := range existing {
		m[f.ID] = f
	}
	return &fakeRepo{byID: m}
}

func (f *fakeRepo) Upsert(ctx context.Context, name string) error { return nil }

func (f *fakeRepo) List(ctx context.Context) ([]domainfeature.Feature, error) {
	out := make([]domainfeature.Feature, 0, len(f.byID))
	for _, v := range f.byID {
		out = append(out, v)
	}
	return out, nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id string) (*domainfeature.Feature, error) {
	v, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	return &v, nil
}

func (f *fakeRepo) Create(ctx context.Context, ft *domainfeature.Feature) error {
	if f.err != nil {
		return f.err
	}
	f.created = ft
	f.byID[ft.ID] = *ft
	return nil
}

func (f *fakeRepo) Delete(ctx context.Context, id string) error {
	if f.err != nil {
		return f.err
	}
	f.deleted = id
	delete(f.byID, id)
	return nil
}

func TestService_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := newFakeRepo()
		svc := feature.NewService(repo, nil)

		f, err := svc.Create(ctx, "Bank")
		require.NoError(t, err)
		require.NotEmpty(t, f.ID)
		require.Equal(t, "Bank", f.Name)
		require.NotNil(t, repo.created)
	})

	t.Run("missing name", func(t *testing.T) {
		repo := newFakeRepo()
		svc := feature.NewService(repo, nil)

		_, err := svc.Create(ctx, "  ")
		require.Error(t, err)
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})

	t.Run("repo error", func(t *testing.T) {
		repo := newFakeRepo()
		repo.err = errors.New("db error")
		svc := feature.NewService(repo, nil)

		_, err := svc.Create(ctx, "Bank")
		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")
	})
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := newFakeRepo(domainfeature.Feature{ID: "feat-1", Name: "Bank"})
		svc := feature.NewService(repo, nil)

		err := svc.Delete(ctx, "feat-1")
		require.NoError(t, err)
		require.Equal(t, "feat-1", repo.deleted)
	})

	t.Run("nonexistent feature", func(t *testing.T) {
		repo := newFakeRepo()
		svc := feature.NewService(repo, nil)

		err := svc.Delete(ctx, "missing")
		require.Error(t, err)
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})
}
