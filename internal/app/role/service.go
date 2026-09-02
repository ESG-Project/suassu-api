package role

import (
	"context"
	"strings"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainrole "github.com/ESG-Project/suassu-api/internal/domain/role"
	"github.com/google/uuid"
)

// ServiceInterface é a porta consumida pela camada HTTP.
type ServiceInterface interface {
	Create(ctx context.Context, enterpriseID, title string) (*domainrole.Role, error)
	List(ctx context.Context, enterpriseID string) ([]domainrole.Role, error)
	Delete(ctx context.Context, roleID, enterpriseID string) error
	ListAssignable(ctx context.Context, actorUserID, enterpriseID string) ([]domainrole.Role, error)
}

type Service struct {
	repo        Repo
	permissions PermissionsByRoleGetter
	users       UserDetails
}

func NewService(r Repo, permissions PermissionsByRoleGetter, users UserDetails) *Service {
	return &Service{repo: r, permissions: permissions, users: users}
}

// Create cadastra um novo papel na empresa do ator.
func (s *Service) Create(ctx context.Context, enterpriseID, title string) (*domainrole.Role, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, apperr.New(apperr.CodeInvalid, "title is required")
	}
	if strings.TrimSpace(enterpriseID) == "" {
		return nil, apperr.New(apperr.CodeInvalid, "enterpriseId is required")
	}

	r := domainrole.NewRole(uuid.NewString(), title, enterpriseID)
	if err := r.Validate(); err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInvalid, "invalid role data")
	}
	if err := s.repo.Create(ctx, r); err != nil {
		return nil, err
	}

	// Auditoria (equivalente ao CreateLogService no user-crud): pendente até o
	// módulo de Logs (Fase 3 da migração) ser portado para o suassu-api.
	return r, nil
}

// List retorna os papéis cadastrados na empresa.
func (s *Service) List(ctx context.Context, enterpriseID string) ([]domainrole.Role, error) {
	return s.repo.List(ctx, enterpriseID)
}

// Delete remove um papel, verificando que ele pertence à empresa do ator.
func (s *Service) Delete(ctx context.Context, roleID, enterpriseID string) error {
	if _, err := s.repo.GetByID(ctx, roleID, enterpriseID); err != nil {
		return apperr.Wrap(err, apperr.CodeNotFound, "role not found")
	}
	return s.repo.Delete(ctx, roleID)
}

// ListAssignable retorna os papéis da empresa que o ator pode atribuir a outro
// usuário: aqueles cujo conjunto de permissões CRUD por feature está contido
// no do próprio ator (anti-escalação de privilégio). O ator sempre pode
// atribuir o próprio papel, mesmo que o cálculo de subconjunto falhasse por
// alguma inconsistência de dados.
func (s *Service) ListAssignable(ctx context.Context, actorUserID, enterpriseID string) ([]domainrole.Role, error) {
	actorDetails, err := s.users.GetUserWithDetails(ctx, actorUserID, enterpriseID)
	if err != nil {
		return nil, err
	}
	actorPerms, err := s.users.GetUserPermissionsWithRole(ctx, actorUserID, enterpriseID)
	if err != nil {
		return nil, err
	}
	actorMatrix := BuildPermissionMatrix(actorPerms.Permissions)

	var actorRoleID string
	if actorDetails.Role != nil {
		actorRoleID = actorDetails.Role.ID
	}

	roles, err := s.repo.List(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	assignable := make([]domainrole.Role, 0, len(roles))
	for _, r := range roles {
		if actorRoleID != "" && r.ID == actorRoleID {
			assignable = append(assignable, r)
			continue
		}

		targetPerms, err := s.permissions.GetByRoleID(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		if IsSubset(targetPerms, actorMatrix) {
			assignable = append(assignable, r)
		}
	}
	return assignable, nil
}

// PermFlags são as flags CRUD de uma permissão numa feature.
type PermFlags struct{ Create, Read, Update, Delete bool }

// BuildPermissionMatrix converte uma lista de permissões num mapa
// featureId -> flags, para comparação eficiente em IsSubset. Exportado para
// reuso pelo módulo de usuários (anti-escalação em criar/editar usuário).
func BuildPermissionMatrix(perms []*types.UserPermission) map[string]PermFlags {
	m := make(map[string]PermFlags, len(perms))
	for _, p := range perms {
		if p == nil {
			continue
		}
		m[p.FeatureID] = PermFlags{p.Create, p.Read, p.Update, p.Delete}
	}
	return m
}

// IsSubset verifica se toda flag ligada em target também está ligada em actor,
// feature a feature — replica utils/rolePermissions.ts isSubset do user-crud.
func IsSubset(target []*types.UserPermission, actor map[string]PermFlags) bool {
	for _, t := range target {
		if t == nil {
			continue
		}
		a := actor[t.FeatureID]
		if t.Create && !a.Create {
			return false
		}
		if t.Read && !a.Read {
			return false
		}
		if t.Update && !a.Update {
			return false
		}
		if t.Delete && !a.Delete {
			return false
		}
	}
	return true
}
