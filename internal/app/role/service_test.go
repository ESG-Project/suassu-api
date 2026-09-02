package role_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ESG-Project/suassu-api/internal/app/role"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainrole "github.com/ESG-Project/suassu-api/internal/domain/role"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	roles   map[string]domainrole.Role // id -> role
	created *domainrole.Role
	deleted string
	err     error
}

func newFakeRepo(roles ...domainrole.Role) *fakeRepo {
	m := make(map[string]domainrole.Role, len(roles))
	for _, r := range roles {
		m[r.ID] = r
	}
	return &fakeRepo{roles: m}
}

func (f *fakeRepo) Create(ctx context.Context, r *domainrole.Role) error {
	if f.err != nil {
		return f.err
	}
	f.created = r
	f.roles[r.ID] = *r
	return nil
}

func (f *fakeRepo) List(ctx context.Context, enterpriseID string) ([]domainrole.Role, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make([]domainrole.Role, 0, len(f.roles))
	for _, r := range f.roles {
		if r.EnterpriseID == enterpriseID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeRepo) GetByID(ctx context.Context, roleID, enterpriseID string) (*types.UserRole, error) {
	r, ok := f.roles[roleID]
	if !ok || r.EnterpriseID != enterpriseID {
		return nil, apperr.New(apperr.CodeNotFound, "role not found")
	}
	return &types.UserRole{ID: r.ID, Title: r.Title}, nil
}

func (f *fakeRepo) Delete(ctx context.Context, roleID string) error {
	if f.err != nil {
		return f.err
	}
	f.deleted = roleID
	delete(f.roles, roleID)
	return nil
}

// fakePermissionsByRole simula PermissionsByRoleGetter: roleID -> permissões.
type fakePermissionsByRole struct {
	byRole map[string][]*types.UserPermission
}

func (f *fakePermissionsByRole) GetByRoleID(ctx context.Context, roleID string) ([]*types.UserPermission, error) {
	return f.byRole[roleID], nil
}

// fakeUserDetails simula UserDetails: actorID -> (role, permissões).
type fakeUserDetails struct {
	roleID string
	perms  []*types.UserPermission
	err    error
}

func (f *fakeUserDetails) GetUserWithDetails(ctx context.Context, userID, enterpriseID string) (*types.UserWithDetails, error) {
	if f.err != nil {
		return nil, f.err
	}
	var role *types.UserRole
	if f.roleID != "" {
		role = &types.UserRole{ID: f.roleID}
	}
	return &types.UserWithDetails{ID: userID, EnterpriseID: enterpriseID, Role: role}, nil
}

func (f *fakeUserDetails) GetUserPermissionsWithRole(ctx context.Context, userID, enterpriseID string) (*types.UserPermissions, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &types.UserPermissions{ID: userID, Permissions: f.perms}, nil
}

func TestService_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := newFakeRepo()
		svc := role.NewService(repo, &fakePermissionsByRole{}, &fakeUserDetails{})

		r, err := svc.Create(ctx, "ent-1", "Técnico")
		require.NoError(t, err)
		require.NotEmpty(t, r.ID)
		require.Equal(t, "Técnico", r.Title)
		require.Equal(t, "ent-1", r.EnterpriseID)
		require.NotNil(t, repo.created)
	})

	t.Run("missing title", func(t *testing.T) {
		repo := newFakeRepo()
		svc := role.NewService(repo, &fakePermissionsByRole{}, &fakeUserDetails{})

		_, err := svc.Create(ctx, "ent-1", "  ")
		require.Error(t, err)
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})

	t.Run("missing enterpriseId", func(t *testing.T) {
		repo := newFakeRepo()
		svc := role.NewService(repo, &fakePermissionsByRole{}, &fakeUserDetails{})

		_, err := svc.Create(ctx, "", "Técnico")
		require.Error(t, err)
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})

	t.Run("repo error", func(t *testing.T) {
		repo := newFakeRepo()
		repo.err = errors.New("db error")
		svc := role.NewService(repo, &fakePermissionsByRole{}, &fakeUserDetails{})

		_, err := svc.Create(ctx, "ent-1", "Técnico")
		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")
	})
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := newFakeRepo(domainrole.Role{ID: "role-1", Title: "Técnico", EnterpriseID: "ent-1"})
		svc := role.NewService(repo, &fakePermissionsByRole{}, &fakeUserDetails{})

		err := svc.Delete(ctx, "role-1", "ent-1")
		require.NoError(t, err)
		require.Equal(t, "role-1", repo.deleted)
	})

	t.Run("role from another enterprise is not found", func(t *testing.T) {
		repo := newFakeRepo(domainrole.Role{ID: "role-1", Title: "Técnico", EnterpriseID: "ent-2"})
		svc := role.NewService(repo, &fakePermissionsByRole{}, &fakeUserDetails{})

		err := svc.Delete(ctx, "role-1", "ent-1")
		require.Error(t, err)
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
		require.Empty(t, repo.deleted)
	})

	t.Run("nonexistent role", func(t *testing.T) {
		repo := newFakeRepo()
		svc := role.NewService(repo, &fakePermissionsByRole{}, &fakeUserDetails{})

		err := svc.Delete(ctx, "missing", "ent-1")
		require.Error(t, err)
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})
}

func TestService_ListAssignable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	adminRole := domainrole.Role{ID: "role-admin", Title: "Administrador", EnterpriseID: "ent-1"}
	techRole := domainrole.Role{ID: "role-tech", Title: "Técnico", EnterpriseID: "ent-1"}
	superRole := domainrole.Role{ID: "role-super", Title: "Super", EnterpriseID: "ent-1"}

	t.Run("actor's own role is always assignable", func(t *testing.T) {
		repo := newFakeRepo(techRole)
		users := &fakeUserDetails{roleID: techRole.ID}
		svc := role.NewService(repo, &fakePermissionsByRole{}, users)

		roles, err := svc.ListAssignable(ctx, "actor-1", "ent-1")
		require.NoError(t, err)
		require.Len(t, roles, 1)
		require.Equal(t, techRole.ID, roles[0].ID)
	})

	t.Run("role whose permissions are a subset of the actor's is assignable", func(t *testing.T) {
		repo := newFakeRepo(adminRole, techRole)
		perms := &fakePermissionsByRole{byRole: map[string][]*types.UserPermission{
			techRole.ID: {{FeatureID: "feat-project", Read: true}},
		}}
		// Ator com papel diferente (admin) e que já possui a flag Read em feat-project.
		users := &fakeUserDetails{
			roleID: adminRole.ID,
			perms:  []*types.UserPermission{{FeatureID: "feat-project", Read: true, Create: true}},
		}
		svc := role.NewService(repo, perms, users)

		roles, err := svc.ListAssignable(ctx, "actor-1", "ent-1")
		require.NoError(t, err)

		ids := make([]string, 0, len(roles))
		for _, r := range roles {
			ids = append(ids, r.ID)
		}
		require.Contains(t, ids, adminRole.ID) // próprio papel do ator
		require.Contains(t, ids, techRole.ID)  // subconjunto do ator
	})

	t.Run("role with more permissions than the actor is excluded (anti-escalation)", func(t *testing.T) {
		repo := newFakeRepo(adminRole, superRole)
		perms := &fakePermissionsByRole{byRole: map[string][]*types.UserPermission{
			superRole.ID: {{FeatureID: "feat-financial", Delete: true}},
		}}
		// Ator não tem Delete em feat-financial.
		users := &fakeUserDetails{
			roleID: adminRole.ID,
			perms:  []*types.UserPermission{{FeatureID: "feat-financial", Read: true}},
		}
		svc := role.NewService(repo, perms, users)

		roles, err := svc.ListAssignable(ctx, "actor-1", "ent-1")
		require.NoError(t, err)

		for _, r := range roles {
			require.NotEqual(t, superRole.ID, r.ID, "superRole não deveria ser atribuível")
		}
	})
}
