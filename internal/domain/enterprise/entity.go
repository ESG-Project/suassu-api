package enterprise

import (
	"errors"
	"strings"

	"github.com/ESG-Project/suassu-api/internal/domain/address"
	"github.com/ESG-Project/suassu-api/internal/domain/parameter"
	"github.com/ESG-Project/suassu-api/internal/domain/product"
	"github.com/ESG-Project/suassu-api/internal/domain/user"
)

type Enterprise struct {
	ID          string
	CNPJ        string
	AddressID   *string
	Address     *address.Address
	Email       string
	FantasyName *string
	Name        string
	Phone       *string
	Users       []*user.User
	Products    []product.Product
	Parameters  []parameter.Parameter
}

func NewEnterprise(id, cnpj, email, name string) *Enterprise {
	return &Enterprise{
		ID:    id,
		CNPJ:  cnpj,
		Email: email,
		Name:  name,
	}
}

func (e *Enterprise) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(e.CNPJ) == "" {
		return errors.New("cnpj is required")
	}
	if strings.TrimSpace(e.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(e.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}

func (e *Enterprise) SetAddressID(addressID *string) {
	e.AddressID = addressID
}

func (e Enterprise) SetAddress(address *address.Address) {
	e.Address = address
}

func (e *Enterprise) SetFantasyName(fantasyName *string) {
	e.FantasyName = fantasyName
}

func (e *Enterprise) SetPhone(phone *string) {
	e.Phone = phone
}

func (e *Enterprise) SetUsers(users []*user.User) {
	e.Users = users
}
