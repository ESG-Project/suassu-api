package adminuser_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ESG-Project/suassu-api/internal/app/adminuser"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainclient "github.com/ESG-Project/suassu-api/internal/domain/client"
	domaintechnician "github.com/ESG-Project/suassu-api/internal/domain/technician"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	postgres "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"

	"github.com/stretchr/testify/require"
)

// ---- fakes ----

type fakeUserRepo struct {
	byID         map[string]*domainuser.User // key = id+"|"+enterpriseId
	byDocument   map[string]*domainuser.User
	byEmail      map[string]*domainuser.User
	primaryAdmin string
	listRows     []postgres.NonClientUserRow
	deletedID    string
	deleteErr    error
}

func key(id, enterpriseID string) string { return id + "|" + enterpriseID }

func (f *fakeUserRepo) GetByID(ctx context.Context, userID, enterpriseID string) (*domainuser.User, error) {
	u, ok := f.byID[key(userID, enterpriseID)]
	if !ok {
		return nil, apperr.New(apperr.CodeNotFound, "user not found")
	}
	return u, nil
}

func (f *fakeUserRepo) GetByEmailForAuth(ctx context.Context, email string) (*domainuser.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, apperr.New(apperr.CodeNotFound, "user not found")
	}
	return u, nil
}

func (f *fakeUserRepo) GetByDocumentInEnterprise(ctx context.Context, enterpriseID, document string) (*domainuser.User, error) {
	u, ok := f.byDocument[document]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (f *fakeUserRepo) GetPrimaryAdminUserID(ctx context.Context, enterpriseID string) (string, error) {
	return f.primaryAdmin, nil
}

func (f *fakeUserRepo) Delete(ctx context.Context, userID, enterpriseID string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deletedID = userID
	return nil
}

func (f *fakeUserRepo) ListNonClientUsersByEnterprise(ctx context.Context, enterpriseID, excludeID string) ([]postgres.NonClientUserRow, error) {
	return f.listRows, nil
}

type fakeRoleRepo struct {
	byID    map[string]*types.UserRole // roleID -> role (already enterprise-scoped)
	byTitle map[string]*types.UserRole // title -> role
}

func (f *fakeRoleRepo) GetByID(ctx context.Context, roleID, enterpriseID string) (*types.UserRole, error) {
	r, ok := f.byID[roleID]
	if !ok {
		return nil, apperr.New(apperr.CodeNotFound, "role not found")
	}
	return r, nil
}

func (f *fakeRoleRepo) GetByTitle(ctx context.Context, enterpriseID, title string) (*types.UserRole, error) {
	r, ok := f.byTitle[title]
	if !ok {
		return nil, apperr.New(apperr.CodeNotFound, "role not found")
	}
	return r, nil
}

type fakePermissionRepo struct {
	byRole map[string][]*types.UserPermission
}

func (f *fakePermissionRepo) GetByRoleID(ctx context.Context, roleID string) ([]*types.UserPermission, error) {
	return f.byRole[roleID], nil
}

type fakeClientRepo struct {
	byUserID map[string]*domainclient.Client
	rows     []postgres.ClientListRow
}

func (f *fakeClientRepo) GetByUserID(ctx context.Context, userID string) (*domainclient.Client, error) {
	return f.byUserID[userID], nil
}
func (f *fakeClientRepo) ListByEnterprise(ctx context.Context, enterpriseID string) ([]postgres.ClientListRow, error) {
	return f.rows, nil
}

type fakeTechnicianRepo struct {
	byUserID map[string]*domaintechnician.Technician
}

func (f *fakeTechnicianRepo) GetByUserID(ctx context.Context, userID string) (*domaintechnician.Technician, error) {
	return f.byUserID[userID], nil
}

type fakeActorPermissions struct {
	details *types.UserWithDetails
	perms   *types.UserPermissions
	err     error
}

func (f *fakeActorPermissions) GetUserWithDetails(ctx context.Context, userID, enterpriseID string) (*types.UserWithDetails, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.details, nil
}
func (f *fakeActorPermissions) GetUserPermissionsWithRole(ctx context.Context, userID, enterpriseID string) (*types.UserPermissions, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.perms, nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(pw string) (string, error) { return "HASH_" + pw, nil }

type fakeTxManager struct {
	called bool
	err    error
}

func (f *fakeTxManager) RunInTx(ctx context.Context, fn func(postgres.Repos) error) error {
	f.called = true
	return f.err
}

func newService(t *testing.T, users *fakeUserRepo, roles *fakeRoleRepo, perms *fakePermissionRepo, clients *fakeClientRepo, techs *fakeTechnicianRepo, actor *fakeActorPermissions, txm *fakeTxManager) *adminuser.Service {
	t.Helper()
	if users == nil {
		users = &fakeUserRepo{byID: map[string]*domainuser.User{}, byDocument: map[string]*domainuser.User{}, byEmail: map[string]*domainuser.User{}}
	}
	if roles == nil {
		roles = &fakeRoleRepo{byID: map[string]*types.UserRole{}, byTitle: map[string]*types.UserRole{}}
	}
	if perms == nil {
		perms = &fakePermissionRepo{byRole: map[string][]*types.UserPermission{}}
	}
	if clients == nil {
		clients = &fakeClientRepo{byUserID: map[string]*domainclient.Client{}}
	}
	if techs == nil {
		techs = &fakeTechnicianRepo{byUserID: map[string]*domaintechnician.Technician{}}
	}
	if actor == nil {
		actor = &fakeActorPermissions{details: &types.UserWithDetails{}, perms: &types.UserPermissions{}}
	}
	if txm == nil {
		txm = &fakeTxManager{}
	}
	return adminuser.NewService(users, roles, perms, clients, techs, actor, fakeHasher{}, txm)
}

// ---- tests ----

func TestCreate_ValidationAndConflicts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	roleTech := &types.UserRole{ID: "role-tech", Title: "Técnico"}
	roles := &fakeRoleRepo{byID: map[string]*types.UserRole{"role-tech": roleTech}, byTitle: map[string]*types.UserRole{"Técnico": roleTech}}

	actor := &fakeActorPermissions{
		details: &types.UserWithDetails{Role: &types.UserRole{ID: "role-tech"}},
		perms:   &types.UserPermissions{Permissions: []*types.UserPermission{{FeatureID: "feat-1", Read: true}}},
	}

	t.Run("missing required fields", func(t *testing.T) {
		txm := &fakeTxManager{}
		svc := newService(t, nil, roles, nil, nil, nil, actor, txm)
		_, err := svc.Create(ctx, "actor-1", "ent-1", adminuser.CreateInput{RoleIDOrTitle: "Técnico"})
		require.Error(t, err)
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
		require.False(t, txm.called)
	})

	t.Run("document already registered", func(t *testing.T) {
		users := &fakeUserRepo{
			byDocument: map[string]*domainuser.User{"123": {}},
			byEmail:    map[string]*domainuser.User{},
		}
		txm := &fakeTxManager{}
		svc := newService(t, users, roles, nil, nil, nil, actor, txm)

		_, err := svc.Create(ctx, "actor-1", "ent-1", adminuser.CreateInput{
			Name: "Ana", Email: "ana@ex.com", Password: "123", Document: "123", RoleIDOrTitle: "Técnico",
		})
		require.Error(t, err)
		require.Equal(t, apperr.CodeConflict, apperr.CodeOf(err))
		require.False(t, txm.called)
	})

	t.Run("email already registered", func(t *testing.T) {
		users := &fakeUserRepo{
			byDocument: map[string]*domainuser.User{},
			byEmail:    map[string]*domainuser.User{"ana@ex.com": {}},
		}
		txm := &fakeTxManager{}
		svc := newService(t, users, roles, nil, nil, nil, actor, txm)

		_, err := svc.Create(ctx, "actor-1", "ent-1", adminuser.CreateInput{
			Name: "Ana", Email: "ana@ex.com", Password: "123", Document: "999", RoleIDOrTitle: "Técnico",
		})
		require.Error(t, err)
		require.Equal(t, apperr.CodeConflict, apperr.CodeOf(err))
		require.False(t, txm.called)
	})

	t.Run("privilege escalation is rejected before touching the transaction", func(t *testing.T) {
		roleAdmin := &types.UserRole{ID: "role-admin", Title: "Administrador"}
		rolesEsc := &fakeRoleRepo{byID: map[string]*types.UserRole{"role-admin": roleAdmin}, byTitle: map[string]*types.UserRole{"Administrador": roleAdmin}}
		perms := &fakePermissionRepo{byRole: map[string][]*types.UserPermission{
			"role-admin": {{FeatureID: "feat-1", Create: true, Read: true, Update: true, Delete: true}},
		}}
		// Ator só tem Read em feat-1, não pode atribuir um papel com Create/Update/Delete.
		lowActor := &fakeActorPermissions{
			details: &types.UserWithDetails{Role: &types.UserRole{ID: "role-tech"}},
			perms:   &types.UserPermissions{Permissions: []*types.UserPermission{{FeatureID: "feat-1", Read: true}}},
		}
		users := &fakeUserRepo{byDocument: map[string]*domainuser.User{}, byEmail: map[string]*domainuser.User{}}
		txm := &fakeTxManager{}
		svc := newService(t, users, rolesEsc, perms, nil, nil, lowActor, txm)

		_, err := svc.Create(ctx, "actor-1", "ent-1", adminuser.CreateInput{
			Name: "Ana", Email: "ana@ex.com", Password: "123", Document: "999", RoleIDOrTitle: "Administrador",
		})
		require.Error(t, err)
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.False(t, txm.called)
	})

	t.Run("transaction error is propagated", func(t *testing.T) {
		users := &fakeUserRepo{byDocument: map[string]*domainuser.User{}, byEmail: map[string]*domainuser.User{}}
		txm := &fakeTxManager{err: errors.New("db exploded")}
		svc := newService(t, users, roles, nil, nil, nil, actor, txm)

		_, err := svc.Create(ctx, "actor-1", "ent-1", adminuser.CreateInput{
			Name: "Ana", Email: "ana@ex.com", Password: "123", Document: "999", RoleIDOrTitle: "Técnico",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "db exploded")
		require.True(t, txm.called)
	})
}

func TestUpdate_Guards(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	roleTech := &types.UserRole{ID: "role-tech", Title: "Técnico"}
	roles := &fakeRoleRepo{byID: map[string]*types.UserRole{"role-tech": roleTech}, byTitle: map[string]*types.UserRole{"Técnico": roleTech}}
	actor := &fakeActorPermissions{
		details: &types.UserWithDetails{Role: &types.UserRole{ID: "role-tech"}},
		perms:   &types.UserPermissions{},
	}

	t.Run("user not found", func(t *testing.T) {
		users := &fakeUserRepo{byID: map[string]*domainuser.User{}}
		txm := &fakeTxManager{}
		svc := newService(t, users, roles, nil, nil, nil, actor, txm)

		_, err := svc.Update(ctx, "actor-1", "ent-1", adminuser.UpdateInput{ID: "missing", RoleIDOrTitle: "Técnico"})
		require.Error(t, err)
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
		require.False(t, txm.called)
	})

	t.Run("primary admin cannot be edited", func(t *testing.T) {
		u := domainuser.NewUser("user-1", "Old", "old@ex.com", "HASH", "123", "ent-1")
		users := &fakeUserRepo{
			byID:         map[string]*domainuser.User{key("user-1", "ent-1"): u},
			primaryAdmin: "user-1",
		}
		txm := &fakeTxManager{}
		svc := newService(t, users, roles, nil, nil, nil, actor, txm)

		_, err := svc.Update(ctx, "actor-1", "ent-1", adminuser.UpdateInput{ID: "user-1", RoleIDOrTitle: "Técnico"})
		require.Error(t, err)
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.False(t, txm.called)
	})

	t.Run("cannot manage a user whose current role exceeds the actor's", func(t *testing.T) {
		roleAdmin := &types.UserRole{ID: "role-admin", Title: "Administrador"}
		rolesEsc := &fakeRoleRepo{
			byID:    map[string]*types.UserRole{"role-tech": roleTech, "role-admin": roleAdmin},
			byTitle: map[string]*types.UserRole{"Técnico": roleTech, "Administrador": roleAdmin},
		}
		perms := &fakePermissionRepo{byRole: map[string][]*types.UserPermission{
			"role-admin": {{FeatureID: "feat-1", Delete: true}},
		}}
		u := domainuser.NewUser("user-1", "Old", "old@ex.com", "HASH", "123", "ent-1")
		adminRoleID := "role-admin"
		u.SetRoleID(&adminRoleID)
		users := &fakeUserRepo{byID: map[string]*domainuser.User{key("user-1", "ent-1"): u}}
		lowActor := &fakeActorPermissions{
			details: &types.UserWithDetails{Role: &types.UserRole{ID: "role-tech"}},
			perms:   &types.UserPermissions{Permissions: []*types.UserPermission{{FeatureID: "feat-1", Read: true}}},
		}
		txm := &fakeTxManager{}
		svc := newService(t, users, rolesEsc, perms, nil, nil, lowActor, txm)

		// Ator tenta manter o mesmo papel (Administrador) do alvo — mas o
		// alvo já tem um papel (Administrador) com mais permissões que o ator.
		_, err := svc.Update(ctx, "actor-1", "ent-1", adminuser.UpdateInput{ID: "user-1", RoleIDOrTitle: "Administrador"})
		require.Error(t, err)
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.False(t, txm.called)
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	u := domainuser.NewUser("user-1", "Ana", "ana@ex.com", "HASH", "123", "ent-1")

	t.Run("success", func(t *testing.T) {
		users := &fakeUserRepo{byID: map[string]*domainuser.User{key("user-1", "ent-1"): u}}
		svc := newService(t, users, nil, nil, nil, nil, nil, nil)

		err := svc.Delete(ctx, "actor-1", "ent-1", "user-1")
		require.NoError(t, err)
		require.Equal(t, "user-1", users.deletedID)
	})

	t.Run("cannot delete yourself", func(t *testing.T) {
		users := &fakeUserRepo{byID: map[string]*domainuser.User{key("user-1", "ent-1"): u}}
		svc := newService(t, users, nil, nil, nil, nil, nil, nil)

		err := svc.Delete(ctx, "user-1", "ent-1", "user-1")
		require.Error(t, err)
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.Empty(t, users.deletedID)
	})

	t.Run("cannot delete the primary admin", func(t *testing.T) {
		users := &fakeUserRepo{byID: map[string]*domainuser.User{key("user-1", "ent-1"): u}, primaryAdmin: "user-1"}
		svc := newService(t, users, nil, nil, nil, nil, nil, nil)

		err := svc.Delete(ctx, "actor-1", "ent-1", "user-1")
		require.Error(t, err)
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.Empty(t, users.deletedID)
	})

	t.Run("not found", func(t *testing.T) {
		users := &fakeUserRepo{byID: map[string]*domainuser.User{}}
		svc := newService(t, users, nil, nil, nil, nil, nil, nil)

		err := svc.Delete(ctx, "actor-1", "ent-1", "missing")
		require.Error(t, err)
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})
}

func TestGetByID_Subtype(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("technician role attaches Technician info", func(t *testing.T) {
		actor := &fakeActorPermissions{details: &types.UserWithDetails{
			ID: "user-1", Role: &types.UserRole{ID: "role-tech", Title: "Técnico"},
		}}
		techs := &fakeTechnicianRepo{byUserID: map[string]*domaintechnician.Technician{
			"user-1": {ID: "tech-1", UserID: "user-1"},
		}}
		svc := newService(t, nil, nil, nil, nil, techs, actor, nil)

		out, err := svc.GetByID(ctx, "ent-1", "user-1")
		require.NoError(t, err)
		require.NotNil(t, out.Technician)
		require.Equal(t, "tech-1", out.Technician.ID)
		require.Nil(t, out.Client)
	})

	t.Run("client role attaches Client info", func(t *testing.T) {
		actor := &fakeActorPermissions{details: &types.UserWithDetails{
			ID: "user-1", Role: &types.UserRole{ID: "role-client", Title: "Cliente"},
		}}
		clients := &fakeClientRepo{byUserID: map[string]*domainclient.Client{
			"user-1": {ID: "client-1", UserID: "user-1"},
		}}
		svc := newService(t, nil, nil, nil, clients, nil, actor, nil)

		out, err := svc.GetByID(ctx, "ent-1", "user-1")
		require.NoError(t, err)
		require.NotNil(t, out.Client)
		require.Equal(t, "client-1", out.Client.ID)
		require.Nil(t, out.Technician)
	})

	t.Run("role without subtype attaches neither", func(t *testing.T) {
		actor := &fakeActorPermissions{details: &types.UserWithDetails{
			ID: "user-1", Role: &types.UserRole{ID: "role-fin", Title: "Financeiro"},
		}}
		svc := newService(t, nil, nil, nil, nil, nil, actor, nil)

		out, err := svc.GetByID(ctx, "ent-1", "user-1")
		require.NoError(t, err)
		require.Nil(t, out.Client)
		require.Nil(t, out.Technician)
	})

	t.Run("user not found", func(t *testing.T) {
		actor := &fakeActorPermissions{err: apperr.New(apperr.CodeNotFound, "user not found")}
		svc := newService(t, nil, nil, nil, nil, nil, actor, nil)

		_, err := svc.GetByID(ctx, "ent-1", "missing")
		require.Error(t, err)
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})
}

func TestListUsers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	title := "Técnico"
	proReg := "PR-123"
	users := &fakeUserRepo{
		listRows: []postgres.NonClientUserRow{
			{ID: "u1", Name: "Ana", Email: "ana@ex.com", Document: "1", RoleTitle: &title, TechnicianID: strPtr("t1"), TechnicianProRegister: &proReg},
			{ID: "u2", Name: "Bia", Email: "bia@ex.com", Document: "2"},
		},
	}
	svc := newService(t, users, nil, nil, nil, nil, nil, nil)

	out, err := svc.ListUsers(ctx, "ent-1")
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.NotNil(t, out[0].Technician)
	require.Equal(t, "PR-123", *out[0].Technician.ProRegister)
	require.Nil(t, out[1].Technician)
}

func TestListClients(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	roleID, roleTitle := "role-client", "Cliente"
	fantasy := "Loja X"
	clients := &fakeClientRepo{rows: []postgres.ClientListRow{
		{ClientID: "c1", UserID: "u1", Name: "Loja", Email: "loja@ex.com", Document: "9", RoleID: &roleID, RoleTitle: &roleTitle, FantasyName: &fantasy},
	}}
	svc := newService(t, nil, nil, nil, clients, nil, nil, nil)

	out, err := svc.ListClients(ctx, "ent-1")
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "c1", out[0].ClientID)
	require.Equal(t, "Cliente", out[0].RoleTitle)
	require.Equal(t, "Loja X", *out[0].FantasyName)
}

func strPtr(s string) *string { return &s }
