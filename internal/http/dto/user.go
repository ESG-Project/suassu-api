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
