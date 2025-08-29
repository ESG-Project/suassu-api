package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/address"
	"github.com/ESG-Project/suassu-api/internal/app/user"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	saved *domainuser.User
	err   error
	users []*domainuser.User
}

func (f *fakeRepo) Create(ctx context.Context, u *domainuser.User) error {
	if f.err != nil {
		return f.err
	}
	f.saved = u
	return nil
}

func (f *fakeRepo) GetByEmailInTenant(ctx context.Context, enterpriseID string, email string) (*domainuser.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if len(f.users) > 0 {
		return f.users[0], nil
	}
	return nil, nil
}

func (f *fakeRepo) GetByEmailForAuth(ctx context.Context, email string) (*domainuser.User, error) {
	return nil, nil
}

func (f *fakeRepo) List(ctx context.Context, enterpriseID string, limit int32, after *domainuser.UserCursorKey) ([]*domainuser.User, domainuser.PageInfo, error) {
	if f.err != nil {
		return nil, domainuser.PageInfo{}, f.err
	}
	return f.users, domainuser.PageInfo{}, nil
}

func (f *fakeRepo) ListAfter(ctx context.Context, enterpriseID string, limit int32, after *domainuser.UserCursorKey) ([]*domainuser.User, domainuser.PageInfo, error) {
	if f.err != nil {
		return nil, domainuser.PageInfo{}, f.err
	}
	return f.users, domainuser.PageInfo{}, nil
}

type fakeHasher struct{ err error }

func (f fakeHasher) Hash(pw string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return "HASH_" + pw, nil
}

func (f fakeHasher) Compare(hash, plain string) error { return nil }

func TestService_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		repo := &fakeRepo{}
		hasher := fakeHasher{}
		svc := user.NewService(repo, &address.Service{}, hasher)

		id, err := svc.Create(ctx, "ent-1", user.CreateInput{
			Name: "Ana", Email: "ana@ex.com", Password: "123",
			Document: "999", EnterpriseID: "ent-1",
		})
		require.NoError(t, err)
		require.NotEmpty(t, id)
		require.NotNil(t, repo.saved)
		require.Equal(t, "HASH_123", repo.saved.PasswordHash)
	})

	t.Run("missing required fields", func(t *testing.T) {
		repo := &fakeRepo{}
		hasher := fakeHasher{}
		svc := user.NewService(repo, &address.Service{}, hasher)

		_, err := svc.Create(ctx, "ent-1", user.CreateInput{
			Name: "", Email: "ana@ex.com", Password: "123", Document: "999", EnterpriseID: "ent-1",
		})
		require.Error(t, err)
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})

	t.Run("missing enterprise ID", func(t *testing.T) {
		repo := &fakeRepo{}
		hasher := fakeHasher{}
		svc := user.NewService(repo, &address.Service{}, hasher)

		_, err := svc.Create(ctx, "", user.CreateInput{
			Name: "Ana", Email: "ana@ex.com", Password: "123", Document: "999", EnterpriseID: "ent-1",
		})
		require.Error(t, err)
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})

	t.Run("hasher error", func(t *testing.T) {
		repo := &fakeRepo{}
		hasher := fakeHasher{err: errors.New("hash failed")}
		svc := user.NewService(repo, &address.Service{}, hasher)

		_, err := svc.Create(ctx, "ent-1", user.CreateInput{
			Name: "Ana", Email: "ana@ex.com", Password: "123", Document: "999", EnterpriseID: "ent-1",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "hash failed")
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &fakeRepo{err: errors.New("db error")}
		hasher := fakeHasher{}
		svc := user.NewService(repo, &address.Service{}, hasher)

		_, err := svc.Create(ctx, "ent-1", user.CreateInput{
			Name: "Ana", Email: "ana@ex.com", Password: "123", Document: "999", EnterpriseID: "ent-1",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")
	})
}

func TestService_List(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		testUsers := []*domainuser.User{
			{ID: "user-1", Name: "Ana", Email: "ana@ex.com", Document: "123", EnterpriseID: "ent-1"},
			{ID: "user-2", Name: "João", Email: "joao@ex.com", Document: "456", EnterpriseID: "ent-1"},
		}
		repo := &fakeRepo{users: testUsers}
		hasher := fakeHasher{}
		svc := user.NewService(repo, &address.Service{}, hasher)

		users, pageInfo, err := svc.List(ctx, "ent-1", 10, nil)
		require.NoError(t, err)
		require.Len(t, users, 2)
		require.Equal(t, "Ana", users[0].Name)
		require.Equal(t, "João", users[1].Name)
		require.NotNil(t, pageInfo)
	})

	t.Run("adjusts invalid limit", func(t *testing.T) {
		repo := &fakeRepo{users: []*domainuser.User{}}
		hasher := fakeHasher{}
		svc := user.NewService(repo, &address.Service{}, hasher)

		users, pageInfo, err := svc.List(ctx, "ent-1", -5, nil)
		require.NoError(t, err)
		require.Len(t, users, 0)
		require.NotNil(t, pageInfo)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &fakeRepo{err: errors.New("db error")}
		hasher := fakeHasher{}
		svc := user.NewService(repo, &address.Service{}, hasher)

		_, _, err := svc.List(ctx, "ent-1", 10, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")
	})
}

func TestService_GetByEmailInTenant(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		testUser := &domainuser.User{ID: "user-1", Name: "Ana", Email: "ana@ex.com", EnterpriseID: "ent-1"}
		repo := &fakeRepo{users: []*domainuser.User{testUser}}
		hasher := fakeHasher{}
		svc := user.NewService(repo, &address.Service{}, hasher)

		user, err := svc.GetByEmailInTenant(ctx, "ent-1", "ana@ex.com")
		require.NoError(t, err)
		require.Equal(t, "Ana", user.Name)
		require.Equal(t, "ana@ex.com", user.Email)
	})

	t.Run("empty email", func(t *testing.T) {
		repo := &fakeRepo{}
		hasher := fakeHasher{}
		svc := user.NewService(repo, &address.Service{}, hasher)

		_, err := svc.GetByEmailInTenant(ctx, "ent-1", "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "email is required")
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &fakeRepo{err: errors.New("db error")}
		hasher := fakeHasher{}
		svc := user.NewService(repo, &address.Service{}, hasher)

		_, err := svc.GetByEmailInTenant(ctx, "ent-1", "ana@ex.com")
		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")
	})
}
