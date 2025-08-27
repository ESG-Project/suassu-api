package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/user"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct{ saved *domainuser.User }

func (f *fakeRepo) Create(ctx context.Context, u *domainuser.User) error { f.saved = u; return nil }
func (f *fakeRepo) GetByEmail(context.Context, string) (*domainuser.User, error) {
	return nil, nil
}
func (f *fakeRepo) List(context.Context, int32, int32) ([]*domainuser.User, error) { return nil, nil }

type fakeHasher struct{}

func (fakeHasher) Hash(pw string) (string, error)   { return "HASH_" + pw, nil }
func (fakeHasher) Compare(hash, plain string) error { return nil }

func TestService_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	repo := &fakeRepo{}
	svc := user.NewService(repo, fakeHasher{})

	id, err := svc.Create(ctx, user.CreateInput{
		Name: "Ana", Email: "ana@ex.com", Password: "123",
		Document: "999", EnterpriseID: "ent-1",
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.NotNil(t, repo.saved)
	require.Equal(t, "HASH_123", repo.saved.PasswordHash)
}
