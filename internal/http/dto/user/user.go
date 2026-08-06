package user

import "github.com/ESG-Project/suassu-api/internal/http/dto/address"

// UserOut representa um usuário na resposta HTTP
type UserOut struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Email        string              `json:"email"`
	Document     string              `json:"document"`
	Phone        *string             `json:"phone,omitempty"`
	AddressID    *string             `json:"addressId,omitempty"`
	Address      *address.AddressOut `json:"address,omitempty"`
	RoleID       *string             `json:"roleId,omitempty"`
	EnterpriseID string              `json:"enterpriseId"`
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

type MeOut struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Email        string              `json:"email"`
	Document     string              `json:"document"`
	Phone        *string             `json:"phone,omitempty"`
	EnterpriseID string              `json:"enterpriseId"`
	Enterprise   *EnterpriseOut      `json:"enterprise,omitempty"`
	AddressID    *string             `json:"addressId,omitempty"`
	Address      *address.AddressOut `json:"address,omitempty"`
	RoleID       *string             `json:"roleId,omitempty"`
	RoleTitle    *string             `json:"role,omitempty"`
	Permissions  []*PermissionOut    `json:"permissions,omitempty"`
}

type MyPermissionsOut struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	RoleTitle   string           `json:"role"`
	Permissions []*PermissionOut `json:"permissions"`
}

// EnterpriseOut representa a empresa do usuário na resposta HTTP
type EnterpriseOut struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	CNPJ        string              `json:"cnpj"`
	Email       string              `json:"email"`
	FantasyName *string             `json:"fantasyName,omitempty"`
	Phone       *string             `json:"phone,omitempty"`
	AddressID   *string             `json:"addressId,omitempty"`
	Address     *address.AddressOut `json:"address,omitempty"`
}
