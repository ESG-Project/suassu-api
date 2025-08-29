package address

import (
	"errors"
	"strings"
)

type Address struct {
	ID           string
	ZipCode      string
	State        string
	City         string
	Neighborhood string
	Street       string
	Num          string
	Latitude     *string
	Longitude    *string
	AddInfo      *string
}

func NewAddress(id, zipCode, state, city, neighborhood, street, num string) *Address {
	return &Address{
		ID:           id,
		ZipCode:      zipCode,
		State:        state,
		City:         city,
		Neighborhood: neighborhood,
		Street:       street,
		Num:          num,
	}
}

func (u *Address) SetLatitude(latitude *string) {
	u.Latitude = latitude
}

func (u *Address) SetLongitude(longitude *string) {
	u.Longitude = longitude
}

func (u *Address) SetAddInfo(addInfo *string) {
	u.AddInfo = addInfo
}

// Validate valida se o endereço está em um estado válido
func (a *Address) Validate() error {
	if strings.TrimSpace(a.ZipCode) == "" {
		return errors.New("zip code is required")
	}
	if strings.TrimSpace(a.State) == "" {
		return errors.New("state is required")
	}
	if strings.TrimSpace(a.City) == "" {
		return errors.New("city is required")
	}
	if strings.TrimSpace(a.Neighborhood) == "" {
		return errors.New("neighborhood is required")
	}
	if strings.TrimSpace(a.Street) == "" {
		return errors.New("street is required")
	}
	if strings.TrimSpace(a.Num) == "" {
		return errors.New("number is required")
	}
	return nil
}

// IsEqual verifica se dois endereços são iguais (para comparação)
func (a *Address) IsEqual(other *Address) bool {
	return a.ZipCode == other.ZipCode &&
		a.State == other.State &&
		a.City == other.City &&
		a.Neighborhood == other.Neighborhood &&
		a.Street == other.Street &&
		a.Num == other.Num
}
