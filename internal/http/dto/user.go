package dto

// AddressOut representa um endereço na resposta HTTP
type AddressOut struct {
	ID           string  `json:"id"`
	State        string  `json:"state"`
	ZipCode      string  `json:"zipCode"`
	City         string  `json:"city"`
	Neighborhood string  `json:"neighborhood"`
	Street       string  `json:"street"`
	Num          string  `json:"num"`
	Latitude     *string `json:"latitude,omitempty"`
	Longitude    *string `json:"longitude,omitempty"`
	AddInfo      *string `json:"addInfo,omitempty"`
}

// UserOut representa um usuário na resposta HTTP
type UserOut struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Email        string      `json:"email"`
	Document     string      `json:"document"`
	Phone        *string     `json:"phone,omitempty"`
	AddressID    *string     `json:"addressId,omitempty"`
	Address      *AddressOut `json:"address,omitempty"`
	RoleID       *string     `json:"roleId,omitempty"`
	EnterpriseID string      `json:"enterpriseId"`
}

type PermissionOut struct {
	ID          string `json:"id"`
	FeatureID   string `json:"featureId"`
	FeatureName string `json:"feature"`
	Create      bool   `json:"create"`
	Read        bool   `json:"read"`
	Update      bool   `json:"update"`
	Delete      bool   `json:"delete"`
}

type EnterpriseOut struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	CNPJ        string      `json:"cnpj"`
	Email       string      `json:"email"`
	FantasyName *string     `json:"fantasyName,omitempty"`
	Phone       *string     `json:"phone,omitempty"`
	AddressID   *string     `json:"addressId,omitempty"`
	Address     *AddressOut `json:"address,omitempty"`
}

type MeOut struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Email        string           `json:"email"`
	Document     string           `json:"document"`
	Phone        *string          `json:"phone,omitempty"`
	EnterpriseID string           `json:"enterpriseId"`
	Enterprise   *EnterpriseOut   `json:"enterprise,omitempty"`
	AddressID    *string          `json:"addressId,omitempty"`
	Address      *AddressOut      `json:"address,omitempty"`
	RoleID       *string          `json:"roleId,omitempty"`
	RoleTitle    *string          `json:"role,omitempty"`
	Permissions  []*PermissionOut `json:"permissions,omitempty"`
}

type MyPermissionsOut struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	RoleTitle   string           `json:"role"`
	Permissions []*PermissionOut `json:"permissions"`
}
