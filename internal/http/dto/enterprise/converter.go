package enterprise

import (
	"github.com/ESG-Project/suassu-api/internal/domain/enterprise"
	"github.com/ESG-Project/suassu-api/internal/http/dto/address"
)

// ToEnterpriseOut converte uma entidade de domínio Enterprise para EnterpriseOut
func ToEnterpriseOut(e *enterprise.Enterprise) *EnterpriseOut {
	if e == nil {
		return nil
	}

	var addrOut *address.AddressOut
	if e.Address != nil {
		addrOut = &address.AddressOut{
			ID:           e.Address.ID,
			ZipCode:      e.Address.ZipCode,
			State:        e.Address.State,
			City:         e.Address.City,
			Neighborhood: e.Address.Neighborhood,
			Street:       e.Address.Street,
			Num:          e.Address.Num,
			Latitude:     e.Address.Latitude,
			Longitude:    e.Address.Longitude,
			AddInfo:      e.Address.AddInfo,
		}
	}

	return &EnterpriseOut{
		ID:          e.ID,
		Name:        e.Name,
		CNPJ:        e.CNPJ,
		Email:       e.Email,
		FantasyName: e.FantasyName,
		Phone:       e.Phone,
		AddressID:   e.AddressID,
		Address:     addrOut,
	}
}
