package enterprise

import "github.com/ESG-Project/suassu-api/internal/http/dto/address"

// EnterpriseOut representa uma empresa na resposta HTTP
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
