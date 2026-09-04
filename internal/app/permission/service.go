package permission

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainpermission "github.com/ESG-Project/suassu-api/internal/domain/permission"
	"github.com/google/uuid"
)

// ServiceInterface é a porta consumida pela camada HTTP.
type ServiceInterface interface {
	Create(ctx context.Context, actorUserID, enterpriseID string, in CreateInput) (*domainpermission.Permission, error)
	Update(ctx context.Context, actorUserID, enterpriseID string, in UpdateInput) (*domainpermission.Permission, error)
	Delete(ctx context.Context, id, enterpriseID string) error
	ListByEnterprise(ctx context.Context, enterpriseID string) ([]domainpermission.Permission, error)
}

type CreateInput struct {
	FeatureID string
	RoleID    string
	Create    bool
	Read      bool
	Update    bool
	Delete    bool
}

type UpdateInput struct {
	ID        string
	FeatureID string
	RoleID    string
	Create    bool
	Read      bool
	Update    bool
	Delete    bool
}

type Service struct {
	repo     Repo
	roles    RoleGetter
	features FeatureGetter
	users    ActorPermissions
}

func NewService(r Repo, roles RoleGetter, features FeatureGetter, users ActorPermissions) *Service {
	return &Service{repo: r, roles: roles, features: features, users: users}
}

// Create cadastra uma permissão (feature x role) para a empresa do ator.
// Replica CreatePermissionService.ts: valida que o role pertence à empresa,
// que a feature existe, e que o ator não está concedendo flags que ele
// próprio não possui (anti-escalação).
func (s *Service) Create(ctx context.Context, actorUserID, enterpriseID string, in CreateInput) (*domainpermission.Permission, error) {
	if in.FeatureID == "" {
		return nil, apperr.New(apperr.CodeInvalid, "featureId is required")
	}
	if in.RoleID == "" {
		return nil, apperr.New(apperr.CodeInvalid, "roleId is required")
	}

	if _, err := s.roles.GetByID(ctx, in.RoleID, enterpriseID); err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "role not found")
	}

	f, err := s.features.GetByID(ctx, in.FeatureID)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, apperr.New(apperr.CodeNotFound, "feature not found")
	}

	if err := s.assertCanGrantFlags(ctx, actorUserID, enterpriseID, in.FeatureID, in.Create, in.Read, in.Update, in.Delete); err != nil {
		return nil, err
	}

	p := domainpermission.NewPermission(uuid.NewString(), in.FeatureID, in.RoleID)
	p.SetPermissions(in.Create, in.Read, in.Update, in.Delete)
	if err := p.Validate(); err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInvalid, "invalid permission data")
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Update altera uma permissão existente, revalidando role/feature/ownership e
// a mesma checagem de anti-escalação usada em Create.
func (s *Service) Update(ctx context.Context, actorUserID, enterpriseID string, in UpdateInput) (*domainpermission.Permission, error) {
	existing, err := s.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if existing.EnterpriseID != enterpriseID {
		return nil, apperr.New(apperr.CodeNotFound, "permission not found")
	}

	if in.FeatureID == "" {
		return nil, apperr.New(apperr.CodeInvalid, "featureId is required")
	}
	if in.RoleID == "" {
		return nil, apperr.New(apperr.CodeInvalid, "roleId is required")
	}

	if _, err := s.roles.GetByID(ctx, in.RoleID, enterpriseID); err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "role not found")
	}

	f, err := s.features.GetByID(ctx, in.FeatureID)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, apperr.New(apperr.CodeNotFound, "feature not found")
	}

	if err := s.assertCanGrantFlags(ctx, actorUserID, enterpriseID, in.FeatureID, in.Create, in.Read, in.Update, in.Delete); err != nil {
		return nil, err
	}

	p := &domainpermission.Permission{
		ID:        existing.ID,
		FeatureID: in.FeatureID,
		RoleID:    in.RoleID,
		Create:    in.Create,
		Read:      in.Read,
		Update:    in.Update,
		Delete:    in.Delete,
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete remove uma permissão, verificando que ela pertence à empresa do ator.
func (s *Service) Delete(ctx context.Context, id, enterpriseID string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.EnterpriseID != enterpriseID {
		return apperr.New(apperr.CodeNotFound, "permission not found")
	}
	return s.repo.Delete(ctx, id)
}

// ListByEnterprise lista todas as permissões da empresa (join por role).
func (s *Service) ListByEnterprise(ctx context.Context, enterpriseID string) ([]domainpermission.Permission, error) {
	return s.repo.ListByEnterprise(ctx, enterpriseID)
}

// assertCanGrantFlags garante que o ator possui, na feature informada, todas
// as flags que está tentando conceder — replica utils/rolePermissions.ts
// assertCanGrantFlags do user-crud. Sem isso, quem tem Permission.create
// conseguiria se autopromover a admin.
func (s *Service) assertCanGrantFlags(ctx context.Context, actorUserID, enterpriseID, featureID string, create, read, update, del bool) error {
	actorPerms, err := s.users.GetUserPermissionsWithRole(ctx, actorUserID, enterpriseID)
	if err != nil {
		return err
	}

	var actorFlags struct{ create, read, update, delete bool }
	for _, p := range actorPerms.Permissions {
		if p != nil && p.FeatureID == featureID {
			actorFlags.create, actorFlags.read, actorFlags.update, actorFlags.delete = p.Create, p.Read, p.Update, p.Delete
			break
		}
	}

	if (create && !actorFlags.create) ||
		(read && !actorFlags.read) ||
		(update && !actorFlags.update) ||
		(del && !actorFlags.delete) {
		return apperr.NewPrivilegeEscalation("you cannot grant permissions you do not have yourself")
	}
	return nil
}
