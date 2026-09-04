package legacyuserdto

import (
	appadmin "github.com/ESG-Project/suassu-api/internal/app/adminuser"
)

// CreateUserRequest é o corpo de POST /user — mesmos campos aceitos pelo
// user-crud (endereço em campos flat, não aninhado).
type CreateUserRequest struct {
	ProRegister  *string `json:"proRegister"`
	Graduation   *string `json:"graduation"`
	CTF          *string `json:"ctf"`
	FantasyName  *string `json:"fantasyName"`
	Name         string  `json:"name"`
	Document     string  `json:"document"`
	Email        string  `json:"email"`
	Password     string  `json:"password"`
	Phone        *string `json:"phone"`
	RoleID       string  `json:"roleId"`
	ZipCode      *string `json:"zipCode"`
	State        *string `json:"state"`
	City         *string `json:"city"`
	Neighborhood *string `json:"neighborhood"`
	Street       *string `json:"street"`
	Num          *string `json:"num"`
	Latitude     *string `json:"latitude"`
	Longitude    *string `json:"longitude"`
	AddInfo      *string `json:"addInfo"`
}

// UpdateUserRequest é o corpo de PUT /user.
type UpdateUserRequest struct {
	ID           string  `json:"id"`
	ProRegister  *string `json:"proRegister"`
	Graduation   *string `json:"graduation"`
	CTF          *string `json:"ctf"`
	FantasyName  *string `json:"fantasyName"`
	Name         string  `json:"name"`
	Document     string  `json:"document"`
	Email        string  `json:"email"`
	Phone        *string `json:"phone"`
	RoleID       string  `json:"roleId"`
	ZipCode      *string `json:"zipCode"`
	State        *string `json:"state"`
	City         *string `json:"city"`
	Neighborhood *string `json:"neighborhood"`
	Street       *string `json:"street"`
	Num          *string `json:"num"`
	Latitude     *string `json:"latitude"`
	Longitude    *string `json:"longitude"`
	AddInfo      *string `json:"addInfo"`
}

// hasAddress reporta se ao menos um campo de endereço foi informado — usado
// para decidir se resolve/cria um Address. Sem isto, um form que não envia
// endereço faria a validação de Address (todos os campos obrigatórios) falhar
// à toa: o user-crud não tem essa restrição (Postgres só exige NOT NULL, que
// string vazia já satisfaz).
func hasAddress(zipCode, state, city, neighborhood, street, num *string) bool {
	for _, f := range []*string{zipCode, state, city, neighborhood, street, num} {
		if f != nil && *f != "" {
			return true
		}
	}
	return false
}

func (r CreateUserRequest) HasAddress() bool {
	return hasAddress(r.ZipCode, r.State, r.City, r.Neighborhood, r.Street, r.Num)
}

func (r UpdateUserRequest) HasAddress() bool {
	return hasAddress(r.ZipCode, r.State, r.City, r.Neighborhood, r.Street, r.Num)
}

// ClientResponse é o sub-registro de Client na resposta.
type ClientResponse struct {
	ID          string  `json:"id"`
	FantasyName *string `json:"fantasyName"`
}

// TechnicianResponse é o sub-registro de Technician na resposta.
type TechnicianResponse struct {
	ID          string  `json:"id"`
	ProRegister *string `json:"proRegister"`
	Graduation  *string `json:"graduation"`
	CTF         *string `json:"ctf"`
}

// UserResponse é a resposta de POST/PUT /user.
type UserResponse struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Email        string              `json:"email"`
	Phone        *string             `json:"phone"`
	Document     string              `json:"document"`
	EnterpriseID string              `json:"enterpriseId"`
	AddressID    *string             `json:"addressId"`
	RoleID       *string             `json:"roleId"`
	Client       *ClientResponse     `json:"client,omitempty"`
	Technician   *TechnicianResponse `json:"technician,omitempty"`
}

func ToUserResponse(u *appadmin.UserOutput) *UserResponse {
	out := &UserResponse{
		ID: u.ID, Name: u.Name, Email: u.Email, Phone: u.Phone,
		Document: u.Document, EnterpriseID: u.EnterpriseID, AddressID: u.AddressID, RoleID: u.RoleID,
	}
	if u.Client != nil {
		out.Client = &ClientResponse{ID: u.Client.ID, FantasyName: u.Client.FantasyName}
	}
	if u.Technician != nil {
		out.Technician = &TechnicianResponse{
			ID: u.Technician.ID, ProRegister: u.Technician.ProRegister,
			Graduation: u.Technician.Graduation, CTF: u.Technician.CTF,
		}
	}
	return out
}

// AddressResponse é o endereço aninhado na resposta de GET /user/:id.
type AddressResponse struct {
	ID           string `json:"id"`
	Street       string `json:"street"`
	Num          string `json:"num"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	ZipCode      string `json:"zipCode"`
}

type EnterpriseRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RoleRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// UserDetailResponse é a resposta de GET /user/:id.
type UserDetailResponse struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Email      string              `json:"email"`
	Phone      *string             `json:"phone"`
	Enterprise EnterpriseRef       `json:"enterprise"`
	Document   string              `json:"document"`
	Address    *AddressResponse    `json:"address"`
	Role       *RoleRef            `json:"role"`
	Client     *ClientResponse     `json:"client,omitempty"`
	Technician *TechnicianResponse `json:"technician,omitempty"`
}

func ToUserDetailResponse(out *appadmin.UserDetailOutput) *UserDetailResponse {
	d := out.Details
	resp := &UserDetailResponse{
		ID: d.ID, Name: d.Name, Email: d.Email, Phone: d.Phone, Document: d.Document,
	}
	if d.Enterprise != nil {
		resp.Enterprise = EnterpriseRef{ID: d.Enterprise.ID, Name: d.Enterprise.Name}
	}
	if d.Address != nil {
		resp.Address = &AddressResponse{
			ID: d.Address.ID, Street: d.Address.Street, Num: d.Address.Num,
			Neighborhood: d.Address.Neighborhood, City: d.Address.City,
			State: d.Address.State, ZipCode: d.Address.ZipCode,
		}
	}
	if d.Role != nil {
		resp.Role = &RoleRef{ID: d.Role.ID, Title: d.Role.Title}
	}
	if out.Client != nil {
		resp.Client = &ClientResponse{ID: out.Client.ID, FantasyName: out.Client.FantasyName}
	}
	if out.Technician != nil {
		resp.Technician = &TechnicianResponse{
			ID: out.Technician.ID, ProRegister: out.Technician.ProRegister,
			Graduation: out.Technician.Graduation, CTF: out.Technician.CTF,
		}
	}
	return resp
}

// UserListItem é um item de GET /user-enterprise.
type UserListItem struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Email        string              `json:"email"`
	Phone        *string             `json:"phone"`
	Document     string              `json:"document"`
	EnterpriseID string              `json:"enterpriseId"`
	Address      *AddressFlatOut     `json:"address"`
	Role         *RoleRef            `json:"role,omitempty"`
	Technician   *TechnicianResponse `json:"technician,omitempty"`
}

// AddressFlatOut é a forma aninhada (sem id) usada nas listagens.
type AddressFlatOut struct {
	Street       *string `json:"street"`
	Num          *string `json:"num"`
	Neighborhood *string `json:"neighborhood"`
	City         *string `json:"city"`
	State        *string `json:"state"`
	ZipCode      *string `json:"zipCode"`
}

func ToUserListItems(list []appadmin.UserOutput) []UserListItem {
	out := make([]UserListItem, 0, len(list))
	for _, u := range list {
		item := UserListItem{
			ID: u.ID, Name: u.Name, Email: u.Email, Phone: u.Phone,
			Document: u.Document, EnterpriseID: u.EnterpriseID,
		}
		if u.RoleID != nil {
			item.Role = &RoleRef{ID: *u.RoleID, Title: u.RoleTitle}
		}
		if u.Technician != nil {
			item.Technician = &TechnicianResponse{
				ID: u.Technician.ID, ProRegister: u.Technician.ProRegister,
				Graduation: u.Technician.Graduation, CTF: u.Technician.CTF,
			}
		}
		if u.Address != nil {
			item.Address = &AddressFlatOut{
				ZipCode: u.Address.ZipCode, State: u.Address.State, City: u.Address.City,
				Neighborhood: u.Address.Neighborhood, Street: u.Address.Street, Num: u.Address.Num,
			}
		}
		out = append(out, item)
	}
	return out
}

// ClientListItemResponse é um item de GET /user/client/enterprise.
type ClientListItemResponse struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Phone        *string        `json:"phone"`
	EnterpriseID string         `json:"enterpriseId"`
	Address      AddressFlatOut `json:"address"`
	Role         RoleRef        `json:"role"`
	ClientID     string         `json:"clientId"`
	FantasyName  *string        `json:"fantasyName"`
	Document     string         `json:"document"`
}

func ToClientListResponses(list []appadmin.ClientListItem) []ClientListItemResponse {
	out := make([]ClientListItemResponse, 0, len(list))
	for _, c := range list {
		out = append(out, ClientListItemResponse{
			ID: c.ID, Name: c.Name, Email: c.Email, Phone: c.Phone,
			EnterpriseID: c.EnterpriseID,
			Address: AddressFlatOut{
				ZipCode: c.Address.ZipCode, State: c.Address.State, City: c.Address.City,
				Neighborhood: c.Address.Neighborhood, Street: c.Address.Street, Num: c.Address.Num,
			},
			Role:        RoleRef{ID: c.RoleID, Title: c.RoleTitle},
			ClientID:    c.ClientID,
			FantasyName: c.FantasyName,
			Document:    c.Document,
		})
	}
	return out
}
