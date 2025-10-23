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

// CreateCompleteRequest representa o request para criar empresa completa com usuário
type CreateCompleteRequest struct {
	Enterprise EnterpriseRequest `json:"enterprise"`
	User       UserRequest       `json:"user"`
}

// EnterpriseRequest representa os dados da empresa no request
type EnterpriseRequest struct {
	Name        string                  `json:"name"`
	CNPJ        string                  `json:"cnpj"`
	Email       string                  `json:"email"`
	FantasyName *string                 `json:"fantasyName,omitempty"`
	Phone       *string                 `json:"phone,omitempty"`
	Address     *address.AddressRequest `json:"address,omitempty"`
}

// UserRequest representa os dados do usuário administrador no request
type UserRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Document string  `json:"document"`
	Phone    *string `json:"phone,omitempty"`
}

// CreateCompleteResponse representa a resposta da criação completa
type CreateCompleteResponse struct {
	EnterpriseID string `json:"enterpriseId"`
	UserID       string `json:"userId"`
	Message      string `json:"message"`
}
