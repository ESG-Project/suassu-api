package enterprise

import (
	"context"
	"fmt"

	"github.com/ESG-Project/suassu-api/internal/app/address"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainenterprise "github.com/ESG-Project/suassu-api/internal/domain/enterprise"
	domainfeature "github.com/ESG-Project/suassu-api/internal/domain/feature"
	domainparameter "github.com/ESG-Project/suassu-api/internal/domain/parameter"
	domainpermission "github.com/ESG-Project/suassu-api/internal/domain/permission"
	domainproduct "github.com/ESG-Project/suassu-api/internal/domain/product"
	domainrole "github.com/ESG-Project/suassu-api/internal/domain/role"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	pg "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
	"github.com/google/uuid"
)

type Service struct {
	repo           Repo
	addressService *address.Service
	hasher         Hasher
	txm            pg.TxManagerInterface
}

func NewService(r Repo, as *address.Service, h Hasher) *Service {
	return NewServiceWithTx(r, as, h, nil)
}

func NewServiceWithTx(r Repo, as *address.Service, h Hasher, txm pg.TxManagerInterface) *Service {
	return &Service{repo: r, addressService: as, hasher: h, txm: txm}
}

type CreateInput struct {
	CNPJ        string
	Email       string
	Name        string
	FantasyName *string
	Phone       *string
	Address     *address.CreateInput
	// Dados do usuário administrador
	User UserInput
}

type UserInput struct {
	Name     string
	Document string
	Email    string
	Password string
	Phone    *string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (enterpriseID string, userID string, err error) {
	if s.txm == nil {
		return "", "", apperr.New(apperr.CodeInternal, "transaction manager not configured")
	}

	// Validações iniciais
	if in.CNPJ == "" || in.Email == "" || in.Name == "" {
		return "", "", apperr.New(apperr.CodeInvalid, "enterprise cnpj, email and name are required")
	}
	if in.User.Name == "" || in.User.Email == "" || in.User.Password == "" || in.User.Document == "" {
		return "", "", apperr.New(apperr.CodeInvalid, "user name, email, password and document are required")
	}

	var createdEnterpriseID, createdUserID string

	err = s.txm.RunInTx(ctx, func(r pg.Repos) error {
		// 1. Criar Address
		var addressID string
		if in.Address != nil {
			addrID, err := s.addressService.HandleAddress(ctx, in.Address)
			if err != nil {
				return fmt.Errorf("failed to create address: %w", err)
			}
			addressID = addrID
		}

		// 2. Criar Enterprise
		enterpriseID := uuid.NewString()
		enterprise := domainenterprise.NewEnterprise(enterpriseID, in.CNPJ, in.Email, in.Name)

		if in.FantasyName != nil {
			enterprise.SetFantasyName(in.FantasyName)
		}
		if in.Phone != nil {
			enterprise.SetPhone(in.Phone)
		}
		if addressID != "" {
			enterprise.SetAddressID(&addressID)
		}

		if err := enterprise.Validate(); err != nil {
			return apperr.Wrap(err, apperr.CodeInvalid, "invalid enterprise data")
		}

		if err := r.Enterprises().Create(ctx, enterprise); err != nil {
			return fmt.Errorf("failed to create enterprise: %w", err)
		}
		createdEnterpriseID = enterpriseID

		// 3. Criar Produto Padrão
		productID := uuid.NewString()
		product := domainproduct.NewProduct(productID, "Combustível", enterpriseID, false)
		suggestedValue := "0"
		product.SetSuggestedValue(&suggestedValue)
		product.SetIsDefault(true)

		if err := product.Validate(); err != nil {
			return apperr.Wrap(err, apperr.CodeInvalid, "invalid product data")
		}

		if err := r.Products().Create(ctx, product); err != nil {
			return fmt.Errorf("failed to create default product: %w", err)
		}

		// 4. Criar Parâmetros Padrão
		defaultParameters := []struct {
			title string
			value string
		}{
			{"Tributo", "0"},
			{"Consumo do automóvel", "0"},
			{"Valor do combustível", "0"},
		}

		for _, param := range defaultParameters {
			paramID := uuid.NewString()
			parameter := domainparameter.NewParameter(paramID, param.title, enterpriseID)
			parameter.SetValue(&param.value)
			parameter.SetIsDefault(true)

			if err := parameter.Validate(); err != nil {
				return apperr.Wrap(err, apperr.CodeInvalid, "invalid parameter data")
			}

			if err := r.Parameters().Create(ctx, parameter); err != nil {
				return fmt.Errorf("failed to create default parameter %s: %w", param.title, err)
			}
		}

		// 5. Criar Roles
		roleIDs := make(map[string]string) // title -> id
		roleNames := []string{"Administrador", "Técnico", "Financeiro", "Cliente"}

		for _, roleName := range roleNames {
			roleID := uuid.NewString()
			role := domainrole.NewRole(roleID, roleName, enterpriseID)

			if err := role.Validate(); err != nil {
				return apperr.Wrap(err, apperr.CodeInvalid, "invalid role data")
			}

			if err := r.Roles().Create(ctx, role); err != nil {
				return fmt.Errorf("failed to create role %s: %w", roleName, err)
			}
			roleIDs[roleName] = roleID
		}

		// 6. Buscar todas as features
		features, err := r.Features().List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list features: %w", err)
		}

		if len(features) == 0 {
			return apperr.New(apperr.CodeInternal, "no features found in database")
		}

		// 7. Criar Permissões
		if err := createPermissionsForRoles(ctx, r, roleIDs, features); err != nil {
			return fmt.Errorf("failed to create permissions: %w", err)
		}

		// 8. Criar Usuário Administrador
		hash, err := s.hasher.Hash(in.User.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		userID := uuid.NewString()
		user := domainuser.NewUser(userID, in.User.Name, in.User.Email, hash, in.User.Document, enterpriseID)

		if in.User.Phone != nil {
			user.SetPhone(in.User.Phone)
		}

		// Atribuir role de Administrador
		adminRoleID := roleIDs["Administrador"]
		user.SetRoleID(&adminRoleID)

		if err := user.Validate(); err != nil {
			return apperr.Wrap(err, apperr.CodeInvalid, "invalid user data")
		}

		// Usar o repositório da transação atual
		if err := r.Users().Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		createdUserID = userID

		return nil
	})

	if err != nil {
		return "", "", err
	}

	return createdEnterpriseID, createdUserID, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*domainenterprise.Enterprise, error) {
	if id == "" {
		return nil, apperr.New(apperr.CodeInvalid, "enterprise id is required")
	}
	return s.repo.GetByID(ctx, id)
}

type UpdateInput struct {
	ID          string
	CNPJ        *string
	Email       *string
	Name        *string
	FantasyName *string
	Phone       *string
	Address     *address.CreateInput
	AddressID   *string
}

func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	enterprise, err := s.repo.GetByID(ctx, in.ID)
	if err != nil {
		return err
	}

	if in.CNPJ != nil {
		enterprise.CNPJ = *in.CNPJ
	}
	if in.Email != nil {
		enterprise.Email = *in.Email
	}
	if in.Name != nil {
		enterprise.Name = *in.Name
	}
	if in.FantasyName != nil {
		enterprise.SetFantasyName(in.FantasyName)
	}
	if in.Phone != nil {
		enterprise.SetPhone(in.Phone)
	}

	if err := enterprise.Validate(); err != nil {
		return apperr.Wrap(err, apperr.CodeInvalid, "invalid enterprise data")
	}

	if in.Address != nil {
		addressID, err := s.addressService.HandleAddress(ctx, in.Address)
		if err != nil {
			return err
		}
		enterprise.SetAddressID(&addressID)
	} else if in.AddressID != nil {
		enterprise.SetAddressID(in.AddressID)
	}

	return s.repo.Update(ctx, enterprise)
}

// createPermissionsForRoles cria permissões para cada role baseado nas regras de negócio
func createPermissionsForRoles(
	ctx context.Context,
	r pg.Repos,
	roleIDs map[string]string,
	features []domainfeature.Feature,
) error {
	// Permissões do Administrador
	adminRoleID := roleIDs["Administrador"]
	for _, feature := range features {
		permID := uuid.NewString()
		perm := domainpermission.NewPermission(permID, feature.ID, adminRoleID)

		// Administrador tem todas as permissões, exceto "erase" em Logs
		if feature.Name == "Logs" {
			perm.SetPermissions(true, true, false, false) // create, read only
		} else {
			perm.SetPermissions(true, true, true, true) // full access
		}

		if err := perm.Validate(); err != nil {
			return err
		}
		if err := r.Permissions().Create(ctx, perm); err != nil {
			return err
		}
	}

	// Permissões do Técnico
	techRoleID := roleIDs["Técnico"]
	restrictedForTech := map[string]bool{
		"Technician":     true,
		"Logs":           true,
		"Parameter":      true,
		"TypeProduct":    true,
		"Product":        true,
		"Transaction":    true,
		"CashFlow":       true,
		"Bank":           true,
		"FinancialCards": true,
	}

	for _, feature := range features {
		permID := uuid.NewString()
		perm := domainpermission.NewPermission(permID, feature.ID, techRoleID)

		if restrictedForTech[feature.Name] {
			if feature.Name == "Logs" {
				perm.SetPermissions(true, false, false, false) // create only
			} else {
				perm.SetPermissions(false, false, false, false) // no access
			}
		} else {
			perm.SetPermissions(true, true, true, true) // full access
		}

		if err := perm.Validate(); err != nil {
			return err
		}
		if err := r.Permissions().Create(ctx, perm); err != nil {
			return err
		}
	}

	// Permissões do Financeiro (sem acesso)
	financialRoleID := roleIDs["Financeiro"]
	for _, feature := range features {
		permID := uuid.NewString()
		perm := domainpermission.NewPermission(permID, feature.ID, financialRoleID)
		perm.SetPermissions(false, false, false, false) // no access

		if err := perm.Validate(); err != nil {
			return err
		}
		if err := r.Permissions().Create(ctx, perm); err != nil {
			return err
		}
	}

	// Permissões do Cliente (sem acesso)
	clientRoleID := roleIDs["Cliente"]
	for _, feature := range features {
		permID := uuid.NewString()
		perm := domainpermission.NewPermission(permID, feature.ID, clientRoleID)
		perm.SetPermissions(false, false, false, false) // no access

		if err := perm.Validate(); err != nil {
			return err
		}
		if err := r.Permissions().Create(ctx, perm); err != nil {
			return err
		}
	}

	return nil
}
