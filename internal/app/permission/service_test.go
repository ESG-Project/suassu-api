package permission_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ESG-Project/suassu-api/internal/app/permission"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainfeature "github.com/ESG-Project/suassu-api/internal/domain/feature"
	domainpermission "github.com/ESG-Project/suassu-api/internal/domain/permission"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	byID    map[string]types.PermissionWithEnterprise
	created *domainpermission.Permission
	updated *domainpermission.Permission
	deleted string
	err     error
}

func newFakeRepo(existing ...types.PermissionWithEnterprise) *fakeRepo {
	m := make(map[string]types.PermissionWithEnterprise, len(existing))
	for _, p := range existing {
		m[p.ID] = p
	}
	return &fakeRepo{byID: m}
}

func (f *fakeRepo) Create(ctx context.Context, p *domainpermission.Permission) error {
	if f.err != nil {
		return f.err
	}
	f.created = p
	return nil
}

func (f *fakeRepo) Update(ctx context.Context, p *domainpermission.Permission) error {
	if f.err != nil {
		return f.err
	}
	f.updated = p
	return nil
}

func (f *fakeRepo) Delete(ctx context.Context, id string) error {
	if f.err != nil {
		return f.err
	}
	f.deleted = id
	return nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id string) (*types.PermissionWithEnterprise, error) {
	p, ok := f.byID[id]
	if !ok {
		return nil, apperr.New(apperr.CodeNotFound, "permission not found")
	}
	return &p, nil
}

func (f *fakeRepo) ListByEnterprise(ctx context.Context, enterpriseID string) ([]domainpermission.Permission, error) {
	var out []domainpermission.Permission
	for _, p := range f.byID {
		if p.EnterpriseID == enterpriseID {
			out = append(out, domainpermission.Permission{ID: p.ID, FeatureID: p.FeatureID, RoleID: p.RoleID})
		}
	}
	return out, nil
}

type fakeRoleGetter struct {
	enterpriseByRole map[string]string
}

func (f *fakeRoleGetter) GetByID(ctx context.Context, roleID, enterpriseID string) (*types.UserRole, error) {
	ent, ok := f.enterpriseByRole[roleID]
	if !ok || ent != enterpriseID {
		return nil, apperr.New(apperr.CodeNotFound, "role not found")
	}
	return &types.UserRole{ID: roleID}, nil
}

type fakeFeatureGetter struct {
	features map[string]bool
}

func (f *fakeFeatureGetter) GetByID(ctx context.Context, id string) (*domainfeature.Feature, error) {
	if !f.features[id] {
		return nil, nil
	}
	return &domainfeature.Feature{ID: id, Name: id}, nil
}

type fakeActorPermissions struct {
	perms []*types.UserPermission
}

func (f *fakeActorPermissions) GetUserPermissionsWithRole(ctx context.Context, userID, enterpriseID string) (*types.UserPermissions, error) {
	return &types.UserPermissions{ID: userID, Permissions: f.perms}, nil
}

func TestService_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	roles := &fakeRoleGetter{enterpriseByRole: map[string]string{"role-1": "ent-1"}}
	features := &fakeFeatureGetter{features: map[string]bool{"feat-1": true}}

	t.Run("success when actor already has the flags being granted", func(t *testing.T) {
		repo := newFakeRepo()
		actor := &fakeActorPermissions{perms: []*types.UserPermission{
			{FeatureID: "feat-1", Create: true, Read: true},
		}}
		svc := permission.NewService(repo, roles, features, actor)

		p, err := svc.Create(ctx, "actor-1", "ent-1", permission.CreateInput{
			FeatureID: "feat-1", RoleID: "role-1", Create: true, Read: true,
		})
		require.NoError(t, err)
		require.NotEmpty(t, p.ID)
		require.NotNil(t, repo.created)
	})

	t.Run("privilege escalation is rejected", func(t *testing.T) {
		repo := newFakeRepo()
		// Ator não tem Delete em feat-1.
		actor := &fakeActorPermissions{perms: []*types.UserPermission{
			{FeatureID: "feat-1", Read: true},
		}}
		svc := permission.NewService(repo, roles, features, actor)

		_, err := svc.Create(ctx, "actor-1", "ent-1", permission.CreateInput{
			FeatureID: "feat-1", RoleID: "role-1", Delete: true,
		})
		require.Error(t, err)
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.Nil(t, repo.created)
	})

	t.Run("role from another enterprise is not found", func(t *testing.T) {
		repo := newFakeRepo()
		actor := &fakeActorPermissions{}
		svc := permission.NewService(repo, roles, features, actor)

		_, err := svc.Create(ctx, "actor-1", "ent-2", permission.CreateInput{
			FeatureID: "feat-1", RoleID: "role-1",
		})
		require.Error(t, err)
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})

	t.Run("nonexistent feature", func(t *testing.T) {
		repo := newFakeRepo()
		actor := &fakeActorPermissions{}
		svc := permission.NewService(repo, roles, features, actor)

		_, err := svc.Create(ctx, "actor-1", "ent-1", permission.CreateInput{
			FeatureID: "missing-feature", RoleID: "role-1",
		})
		require.Error(t, err)
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})

	t.Run("missing featureId", func(t *testing.T) {
		repo := newFakeRepo()
		actor := &fakeActorPermissions{}
		svc := permission.NewService(repo, roles, features, actor)

		_, err := svc.Create(ctx, "actor-1", "ent-1", permission.CreateInput{RoleID: "role-1"})
		require.Error(t, err)
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})

	t.Run("repo error", func(t *testing.T) {
		repo := newFakeRepo()
		repo.err = errors.New("db error")
		actor := &fakeActorPermissions{}
		svc := permission.NewService(repo, roles, features, actor)

		_, err := svc.Create(ctx, "actor-1", "ent-1", permission.CreateInput{
			FeatureID: "feat-1", RoleID: "role-1",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")
	})
}

func TestService_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	roles := &fakeRoleGetter{enterpriseByRole: map[string]string{"role-1": "ent-1"}}
	features := &fakeFeatureGetter{features: map[string]bool{"feat-1": true}}
	existing := types.PermissionWithEnterprise{ID: "perm-1", FeatureID: "feat-1", RoleID: "role-1", EnterpriseID: "ent-1"}

	t.Run("success", func(t *testing.T) {
		repo := newFakeRepo(existing)
		actor := &fakeActorPermissions{perms: []*types.UserPermission{{FeatureID: "feat-1", Update: true}}}
		svc := permission.NewService(repo, roles, features, actor)

		p, err := svc.Update(ctx, "actor-1", "ent-1", permission.UpdateInput{
			ID: "perm-1", FeatureID: "feat-1", RoleID: "role-1", Update: true,
		})
		require.NoError(t, err)
		require.True(t, p.Update)
		require.NotNil(t, repo.updated)
	})

	t.Run("permission from another enterprise is not found", func(t *testing.T) {
		repo := newFakeRepo(existing)
		actor := &fakeActorPermissions{}
		svc := permission.NewService(repo, roles, features, actor)

		_, err := svc.Update(ctx, "actor-1", "ent-2", permission.UpdateInput{
			ID: "perm-1", FeatureID: "feat-1", RoleID: "role-1",
		})
		require.Error(t, err)
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})

	t.Run("privilege escalation is rejected", func(t *testing.T) {
		repo := newFakeRepo(existing)
		actor := &fakeActorPermissions{} // sem nenhuma flag
		svc := permission.NewService(repo, roles, features, actor)

		_, err := svc.Update(ctx, "actor-1", "ent-1", permission.UpdateInput{
			ID: "perm-1", FeatureID: "feat-1", RoleID: "role-1", Delete: true,
		})
		require.Error(t, err)
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
	})
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	existing := types.PermissionWithEnterprise{ID: "perm-1", EnterpriseID: "ent-1"}

	t.Run("success", func(t *testing.T) {
		repo := newFakeRepo(existing)
		svc := permission.NewService(repo, &fakeRoleGetter{}, &fakeFeatureGetter{}, &fakeActorPermissions{})

		err := svc.Delete(ctx, "perm-1", "ent-1")
		require.NoError(t, err)
		require.Equal(t, "perm-1", repo.deleted)
	})

	t.Run("permission from another enterprise is not found", func(t *testing.T) {
		repo := newFakeRepo(existing)
		svc := permission.NewService(repo, &fakeRoleGetter{}, &fakeFeatureGetter{}, &fakeActorPermissions{})

		err := svc.Delete(ctx, "perm-1", "ent-2")
		require.Error(t, err)
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
		require.Empty(t, repo.deleted)
	})
}
