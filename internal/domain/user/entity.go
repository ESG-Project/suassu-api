package user

import (
	"errors"
	"strings"

	"github.com/ESG-Project/suassu-api/internal/domain/address"
)

// User representa a entidade de usuário no domínio
type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Document     string
	Phone        *string
	AddressID    *string
	Address      *address.Address
	RoleID       *string
	EnterpriseID string
	Active       bool
}

// NewUser cria uma nova instância de User
func NewUser(id, name, email, passwordHash, document, enterpriseID string) *User {
	return &User{
		ID:           id,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Document:     document,
		EnterpriseID: enterpriseID,
		Active:       true,
	}
}

// Validate valida se o usuário está em um estado válido
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(u.Document) == "" {
		return errors.New("document is required")
	}
	if strings.TrimSpace(u.EnterpriseID) == "" {
		return errors.New("enterprise ID is required")
	}
	if strings.TrimSpace(u.PasswordHash) == "" {
		return errors.New("password hash is required")
	}
	return nil
}

// Activate ativa o usuário
func (u *User) Activate() {
	u.Active = true
}

// Deactivate desativa o usuário
func (u *User) Deactivate() {
	u.Active = false
}

// IsActive verifica se o usuário está ativo
func (u *User) IsActive() bool {
	return u.Active
}

// SetPhone define o telefone do usuário
func (u *User) SetPhone(phone *string) {
	u.Phone = phone
}

// SetAddressID define o ID do endereço do usuário
func (u *User) SetAddressID(addressID *string) {
	u.AddressID = addressID
}

func (u *User) SetAddress(address *address.Address) {
	u.Address = address
}

// SetRoleID define o ID do papel do usuário
func (u *User) SetRoleID(roleID *string) {
	u.RoleID = roleID
}
