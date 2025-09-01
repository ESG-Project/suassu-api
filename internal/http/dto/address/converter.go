package address

import "github.com/ESG-Project/suassu-api/internal/domain/address"

// ToAddressOut converte uma entidade de domínio Address para AddressOut
func ToAddressOut(addr *address.Address) *AddressOut {
	if addr == nil {
		return nil
	}

	return &AddressOut{
		ID:           addr.ID,
		State:        addr.State,
		ZipCode:      addr.ZipCode,
		City:         addr.City,
		Neighborhood: addr.Neighborhood,
		Street:       addr.Street,
		Num:          addr.Num,
		Latitude:     addr.Latitude,
		Longitude:    addr.Longitude,
		AddInfo:      addr.AddInfo,
	}
}
