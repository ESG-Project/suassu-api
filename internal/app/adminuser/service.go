package adminuser

import (
	"context"
	"strings"

	"github.com/ESG-Project/suassu-api/internal/app/address"
	approle "github.com/ESG-Project/suassu-api/internal/app/role"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainclient "github.com/ESG-Project/suassu-api/internal/domain/client"
	domaintechnician "github.com/ESG-Project/suassu-api/internal/domain/technician"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	postgres "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
	"github.com/google/uuid"
)

// Títulos de papel que disparam a criação do sub-registro correspondente.
// Espelha utils/roleKind.ts do user-crud: comparação exata de propósito, para
// preservar o comportamento atual — não é (ainda) um discriminador explícito
// no schema.
const (
	technicianRoleTitle = "Técnico"
	clientRoleTitle     = "Cliente"
)

func requiresTechnicianRecord(title string) bool { return title == technicianRoleTitle }
func requiresClientRecord(title string) bool     { return title == clientRoleTitle }

type Hasher interface {
	Hash(pw string) (string, error)
}

type Service struct {
	users       UserRepo
	roles       RoleRepo
	permissions PermissionRepo
	clients     ClientRepo
	technicians TechnicianRepo
	actorSvc    ActorPermissions
	hasher      Hasher
	txm         postgres.TxManagerInterface
}

// NewService monta o serviço administrativo de usuários. A resolução de
// endereço é feita dentro da transação (ver Create/Update), por isso não
// recebe um *address.Service pronto: um novo é montado por chamada, com o
// repositório escopado à transação corrente (r.Addresses()).
func NewService(
	users UserRepo,
	roles RoleRepo,
	permissions PermissionRepo,
	clients ClientRepo,
	technicians TechnicianRepo,
	actorSvc ActorPermissions,
	hasher Hasher,
	txm postgres.TxManagerInterface,
) *Service {
	return &Service{
		users:       users,
		roles:       roles,
		permissions: permissions,
		clients:     clients,
		technicians: technicians,
		actorSvc:    actorSvc,
		hasher:      hasher,
		txm:         txm,
	}
}

// ClientInfo é o sub-registro de Client de um usuário, na saída da API.
type ClientInfo struct {
	ID          string
	FantasyName *string
}

// TechnicianInfo é o sub-registro de Technician de um usuário, na saída da API.
type TechnicianInfo struct {
	ID          string
	ProRegister *string
	Graduation  *string
	CTF         *string
}

// UserOutput é a forma comum de retorno de Create/Update/List — o usuário
// mais, quando aplicável, seu sub-registro de Client ou Technician.
type UserOutput struct {
	ID           string
	Name         string
	Email        string
	Phone        *string
	Document     string
	EnterpriseID string
	AddressID    *string
	Address      *AddressOut
	RoleID       *string
	RoleTitle    string
	Client       *ClientInfo
	Technician   *TechnicianInfo
}

// UserDetailOutput é a saída de GetByID ("findOne"): o usuário com detalhes
// completos (endereço, papel, empresa) mais o sub-registro de Client ou
// Technician quando aplicável.
type UserDetailOutput struct {
	Details    *types.UserWithDetails
	Client     *ClientInfo
	Technician *TechnicianInfo
}

// resolveRole aceita tanto um UUID de papel quanto o título em português
// ("Cliente", "Técnico", "Financeiro" etc.) — mesmo comportamento do
// CreateUserService/UpdateUserService no user-crud.
func (s *Service) resolveRole(ctx context.Context, enterpriseID, roleIDOrTitle string) (*types.UserRole, error) {
	roleIDOrTitle = strings.TrimSpace(roleIDOrTitle)
	if roleIDOrTitle == "" {
		return nil, apperr.New(apperr.CodeInvalid, "roleId is required")
	}

	if looksLikeUUID(roleIDOrTitle) {
		r, err := s.roles.GetByID(ctx, roleIDOrTitle, enterpriseID)
		if err != nil {
			return nil, apperr.Wrap(err, apperr.CodeNotFound, "role not found")
		}
		return r, nil
	}

	r, err := s.roles.GetByTitle(ctx, enterpriseID, roleIDOrTitle)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "role not found")
	}
	return r, nil
}

func looksLikeUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// actorMatrix resolve a matriz de permissões e o roleId do ator, para as
// checagens de anti-escalação.
func (s *Service) actorMatrix(ctx context.Context, actorUserID, enterpriseID string) (map[string]approle.PermFlags, string, error) {
	details, err := s.actorSvc.GetUserWithDetails(ctx, actorUserID, enterpriseID)
	if err != nil {
		return nil, "", err
	}
	perms, err := s.actorSvc.GetUserPermissionsWithRole(ctx, actorUserID, enterpriseID)
	if err != nil {
		return nil, "", err
	}

	var actorRoleID string
	if details.Role != nil {
		actorRoleID = details.Role.ID
	}
	return approle.BuildPermissionMatrix(perms.Permissions), actorRoleID, nil
}

// assertCanAssignRole garante que o ator só atribui papéis cujo conjunto de
// permissões está contido no dele próprio (anti-escalação). O ator sempre
// pode atribuir o próprio papel.
func (s *Service) assertCanAssignRole(ctx context.Context, target *types.UserRole, actorMatrix map[string]approle.PermFlags, actorRoleID string) error {
	if actorRoleID != "" && actorRoleID == target.ID {
		return nil
	}
	targetPerms, err := s.permissions.GetByRoleID(ctx, target.ID)
	if err != nil {
		return err
	}
	if !approle.IsSubset(targetPerms, actorMatrix) {
		return apperr.NewPrivilegeEscalation("você não pode atribuir o cargo \"" + target.Title + "\" porque ele possui permissões que você não tem")
	}
	return nil
}

// assertCanManageUserWithRole garante que o ator só edita/gerencia usuários
// cujo papel ATUAL está contido no dele próprio — sem isto, quem tem
// User.update conseguiria rebaixar ou sequestrar a conta de alguém com mais
// permissões do que ele.
func (s *Service) assertCanManageUserWithRole(ctx context.Context, currentRoleID, enterpriseID string, actorMatrix map[string]approle.PermFlags, actorRoleID string) error {
	if actorRoleID != "" && actorRoleID == currentRoleID {
		return nil
	}
	currentRole, err := s.roles.GetByID(ctx, currentRoleID, enterpriseID)
	if err != nil {
		return apperr.Wrap(err, apperr.CodeNotFound, "role not found")
	}
	targetPerms, err := s.permissions.GetByRoleID(ctx, currentRoleID)
	if err != nil {
		return err
	}
	if !approle.IsSubset(targetPerms, actorMatrix) {
		return apperr.NewPrivilegeEscalation("você não pode editar um usuário com o cargo \"" + currentRole.Title + "\", que possui permissões que você não tem")
	}
	return nil
}

// assertNotPrimaryAdmin bloqueia edição/exclusão do usuário administrador
// "primário" da empresa (criado no onboarding).
func (s *Service) assertNotPrimaryAdmin(ctx context.Context, enterpriseID, targetUserID string) error {
	primaryID, err := s.users.GetPrimaryAdminUserID(ctx, enterpriseID)
	if err != nil {
		return err
	}
	if primaryID != "" && primaryID == targetUserID {
		return apperr.New(apperr.CodeForbidden, "o usuário administrador primário não pode ser alterado")
	}
	return nil
}

// CreateInput é a entrada de Create.
type CreateInput struct {
	Name          string
	Document      string
	Email         string
	Password      string
	Phone         *string
	RoleIDOrTitle string
	Address       *address.CreateInput
	ProRegister   *string
	Graduation    *string
	CTF           *string
	FantasyName   *string
}

// Create cadastra um usuário administrativamente (distinto do self-service),
// criando o sub-registro de Client/Technician quando o papel exigir.
// Replica CreateUserService.execute do user-crud.
func (s *Service) Create(ctx context.Context, actorUserID, enterpriseID string, in CreateInput) (*UserOutput, error) {
	if in.Name == "" || in.Email == "" || in.Password == "" || in.Document == "" {
		return nil, apperr.New(apperr.CodeInvalid, "name, email, password and document are required")
	}

	role, err := s.resolveRole(ctx, enterpriseID, in.RoleIDOrTitle)
	if err != nil {
		return nil, err
	}

	actorMatrix, actorRoleID, err := s.actorMatrix(ctx, actorUserID, enterpriseID)
	if err != nil {
		return nil, err
	}
	if err := s.assertCanAssignRole(ctx, role, actorMatrix, actorRoleID); err != nil {
		return nil, err
	}

	if existing, err := s.users.GetByDocumentInEnterprise(ctx, enterpriseID, in.Document); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, apperr.New(apperr.CodeConflict, "já existe um usuário com esse documento registrado")
	}
	if existing, err := s.users.GetByEmailForAuth(ctx, in.Email); err != nil && apperr.CodeOf(err) != apperr.CodeNotFound {
		return nil, err
	} else if existing != nil {
		return nil, apperr.New(apperr.CodeConflict, "este email já está em uso")
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	userID := uuid.NewString()
	u := domainuser.NewUser(userID, in.Name, in.Email, hash, in.Document, enterpriseID)
	if in.Phone != nil {
		u.SetPhone(in.Phone)
	}
	u.SetRoleID(&role.ID)

	var out *UserOutput
	err = s.txm.RunInTx(ctx, func(r postgres.Repos) error {
		if in.Address != nil {
			addrSvc := address.NewService(r.Addresses(), nil)
			addressID, err := addrSvc.HandleAddress(ctx, in.Address)
			if err != nil {
				return err
			}
			u.SetAddressID(&addressID)
		}

		if err := u.Validate(); err != nil {
			return apperr.Wrap(err, apperr.CodeInvalid, "invalid user data")
		}
		if err := r.Users().Create(ctx, u); err != nil {
			return err
		}

		out = &UserOutput{
			ID: u.ID, Name: u.Name, Email: u.Email, Phone: u.Phone,
			Document: u.Document, EnterpriseID: u.EnterpriseID, AddressID: u.AddressID,
			RoleID: u.RoleID, RoleTitle: role.Title,
		}

		switch {
		case requiresTechnicianRecord(role.Title):
			t := domaintechnician.NewTechnician(uuid.NewString(), userID)
			t.SetProRegister(in.ProRegister)
			t.SetGraduation(in.Graduation)
			t.SetCTF(in.CTF)
			if err := r.Technicians().Create(ctx, t); err != nil {
				return err
			}
			out.Technician = &TechnicianInfo{ID: t.ID, ProRegister: t.ProRegister, Graduation: t.Graduation, CTF: t.CTF}
		case requiresClientRecord(role.Title):
			c := domainclient.NewClient(uuid.NewString(), userID)
			c.SetFantasyName(in.FantasyName)
			if err := r.Clients().Create(ctx, c); err != nil {
				return err
			}
			out.Client = &ClientInfo{ID: c.ID, FantasyName: c.FantasyName}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CreateMany cadastra vários usuários sequencialmente, parando no primeiro
// erro (índice retornado é 0-based). Replica CreateUserService.many, exceto
// pela concorrência: lá cada usuário é criado com Promise.all (paralelo, sem
// garantia de atomicidade entre eles); aqui é sequencial, o que é mais fácil
// de raciocinar e produz o mesmo resultado prático (sucesso parcial até o
// primeiro erro).
func (s *Service) CreateMany(ctx context.Context, actorUserID, enterpriseID string, items []CreateInput) error {
	for _, item := range items {
		if strings.TrimSpace(item.Email) == "" && item.Document != "" {
			item.Email = item.Document + "@suassu.com"
		}
		if _, err := s.Create(ctx, actorUserID, enterpriseID, item); err != nil {
			return err
		}
	}
	return nil
}

// UpdateInput é a entrada de Update.
type UpdateInput struct {
	ID            string
	Name          string
	Document      string
	Email         string
	Phone         *string
	RoleIDOrTitle string
	Address       *address.CreateInput
	ProRegister   *string
	Graduation    *string
	CTF           *string
	FantasyName   *string
}

// Update edita um usuário administrativamente. Replica UpdateUserService.updateUser.
func (s *Service) Update(ctx context.Context, actorUserID, enterpriseID string, in UpdateInput) (*UserOutput, error) {
	existing, err := s.users.GetByID(ctx, in.ID, enterpriseID)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "user not found")
	}

	if err := s.assertNotPrimaryAdmin(ctx, enterpriseID, in.ID); err != nil {
		return nil, err
	}

	role, err := s.resolveRole(ctx, enterpriseID, in.RoleIDOrTitle)
	if err != nil {
		return nil, err
	}

	actorMatrix, actorRoleID, err := s.actorMatrix(ctx, actorUserID, enterpriseID)
	if err != nil {
		return nil, err
	}
	if err := s.assertCanAssignRole(ctx, role, actorMatrix, actorRoleID); err != nil {
		return nil, err
	}
	if existing.RoleID != nil {
		if err := s.assertCanManageUserWithRole(ctx, *existing.RoleID, enterpriseID, actorMatrix, actorRoleID); err != nil {
			return nil, err
		}
	}

	existing.Name = in.Name
	existing.Email = in.Email
	existing.Document = in.Document
	existing.SetPhone(in.Phone)
	existing.SetRoleID(&role.ID)

	var out *UserOutput
	err = s.txm.RunInTx(ctx, func(r postgres.Repos) error {
		if in.Address != nil {
			addrSvc := address.NewService(r.Addresses(), nil)
			addressID, err := addrSvc.HandleAddress(ctx, in.Address)
			if err != nil {
				return err
			}
			existing.SetAddressID(&addressID)
		}

		if err := existing.Validate(); err != nil {
			return apperr.Wrap(err, apperr.CodeInvalid, "invalid user data")
		}
		if err := r.Users().AdminUpdate(ctx, existing); err != nil {
			return err
		}

		out = &UserOutput{
			ID: existing.ID, Name: existing.Name, Email: existing.Email, Phone: existing.Phone,
			Document: existing.Document, EnterpriseID: existing.EnterpriseID, AddressID: existing.AddressID,
			RoleID: existing.RoleID, RoleTitle: role.Title,
		}

		switch {
		case requiresTechnicianRecord(role.Title):
			t := domaintechnician.NewTechnician(uuid.NewString(), existing.ID)
			t.SetProRegister(in.ProRegister)
			t.SetGraduation(in.Graduation)
			t.SetCTF(in.CTF)
			if err := r.Technicians().Upsert(ctx, t); err != nil {
				return err
			}
			out.Technician = &TechnicianInfo{ID: t.ID, ProRegister: t.ProRegister, Graduation: t.Graduation, CTF: t.CTF}
		case requiresClientRecord(role.Title):
			c := domainclient.NewClient(uuid.NewString(), existing.ID)
			c.SetFantasyName(in.FantasyName)
			if err := r.Clients().Upsert(ctx, c); err != nil {
				return err
			}
			out.Client = &ClientInfo{ID: c.ID, FantasyName: c.FantasyName}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Delete remove um usuário administrativamente, bloqueando auto-exclusão e
// exclusão do admin primário. Replica DeleteUserService.
func (s *Service) Delete(ctx context.Context, actorUserID, enterpriseID, targetUserID string) error {
	if _, err := s.users.GetByID(ctx, targetUserID, enterpriseID); err != nil {
		return apperr.Wrap(err, apperr.CodeNotFound, "user not found")
	}
	if targetUserID == actorUserID {
		return apperr.New(apperr.CodeForbidden, "você não pode deletar o seu próprio usuário")
	}
	if err := s.assertNotPrimaryAdmin(ctx, enterpriseID, targetUserID); err != nil {
		return err
	}
	return s.users.Delete(ctx, targetUserID, enterpriseID)
}

// GetByID busca um usuário com detalhes completos (endereço, papel, empresa)
// e o sub-registro de Client/Technician quando aplicável. Replica
// ReadUserService.one.
func (s *Service) GetByID(ctx context.Context, enterpriseID, targetUserID string) (*UserDetailOutput, error) {
	details, err := s.getDetails(ctx, enterpriseID, targetUserID)
	if err != nil {
		return nil, err
	}

	out := &UserDetailOutput{Details: details}
	if details.Role == nil {
		return out, nil
	}

	switch {
	case requiresTechnicianRecord(details.Role.Title):
		t, err := s.technicians.GetByUserID(ctx, targetUserID)
		if err == nil && t != nil {
			out.Technician = &TechnicianInfo{ID: t.ID, ProRegister: t.ProRegister, Graduation: t.Graduation, CTF: t.CTF}
		}
	case requiresClientRecord(details.Role.Title):
		c, err := s.clients.GetByUserID(ctx, targetUserID)
		if err == nil && c != nil {
			out.Client = &ClientInfo{ID: c.ID, FantasyName: c.FantasyName}
		}
	}
	return out, nil
}

func (s *Service) getDetails(ctx context.Context, enterpriseID, targetUserID string) (*types.UserWithDetails, error) {
	details, err := s.actorSvc.GetUserWithDetails(ctx, targetUserID, enterpriseID)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "user not found")
	}
	return details, nil
}

// ListUsers lista os usuários da empresa que não têm registro de Client
// (equivalente a manyUsers no user-crud), excluindo o admin primário.
func (s *Service) ListUsers(ctx context.Context, enterpriseID string) ([]UserOutput, error) {
	primaryID, err := s.users.GetPrimaryAdminUserID(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	rows, err := s.users.ListNonClientUsersByEnterprise(ctx, enterpriseID, primaryID)
	if err != nil {
		return nil, err
	}

	out := make([]UserOutput, 0, len(rows))
	for _, row := range rows {
		u := UserOutput{
			ID: row.ID, Name: row.Name, Email: row.Email, Phone: row.Phone,
			Document: row.Document, EnterpriseID: row.EnterpriseID,
			RoleID: row.RoleID,
		}
		if row.ZipCode != nil || row.State != nil || row.City != nil || row.Neighborhood != nil || row.Street != nil || row.Num != nil {
			u.Address = &AddressOut{
				ZipCode: row.ZipCode, State: row.State, City: row.City,
				Neighborhood: row.Neighborhood, Street: row.Street, Num: row.Num,
			}
		}
		if row.RoleTitle != nil {
			u.RoleTitle = *row.RoleTitle
		}
		if row.TechnicianID != nil {
			u.Technician = &TechnicianInfo{
				ID: *row.TechnicianID, ProRegister: row.TechnicianProRegister,
				Graduation: row.TechnicianGraduation, CTF: row.TechnicianCTF,
			}
		}
		out = append(out, u)
	}
	return out, nil
}

// ListClients lista os usuários com papel "Cliente" (equivalente a
// manyClients no user-crud).
func (s *Service) ListClients(ctx context.Context, enterpriseID string) ([]ClientListItem, error) {
	rows, err := s.clients.ListByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	out := make([]ClientListItem, 0, len(rows))
	for _, row := range rows {
		item := ClientListItem{
			ID: row.UserID, ClientID: row.ClientID, Name: row.Name, Email: row.Email,
			Phone: row.Phone, Document: row.Document, EnterpriseID: row.EnterpriseID,
			FantasyName: row.FantasyName,
			Address: AddressOut{
				ZipCode: row.ZipCode, State: row.State, City: row.City,
				Neighborhood: row.Neighborhood, Street: row.Street, Num: row.Num,
			},
		}
		if row.RoleID != nil {
			item.RoleID = *row.RoleID
		}
		if row.RoleTitle != nil {
			item.RoleTitle = *row.RoleTitle
		}
		out = append(out, item)
	}
	return out, nil
}

// AddressOut é a forma flat de endereço usada nas listagens de client/technician.
type AddressOut struct {
	ZipCode, State, City, Neighborhood, Street, Num *string
}

// ClientListItem é um item de ListClients.
type ClientListItem struct {
	ID           string
	ClientID     string
	Name         string
	Email        string
	Phone        *string
	Document     string
	EnterpriseID string
	FantasyName  *string
	RoleID       string
	RoleTitle    string
	Address      AddressOut
}
