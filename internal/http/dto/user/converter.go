package user

import (
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/http/dto/address"
)

// ToUserOut converte uma entidade de domínio User para UserOut
func ToUserOut(u *user.User) *UserOut {
	if u == nil {
		return nil
	}

	var addrOut *address.AddressOut
	if u.Address != nil {
		addrOut = address.ToAddressOut(u.Address)
	}

	return &UserOut{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		Document:     u.Document,
		Phone:        u.Phone,
		AddressID:    u.AddressID,
		Address:      addrOut,
		RoleID:       u.RoleID,
		EnterpriseID: u.EnterpriseID,
	}
}

// ToMeOut converte um UserWithDetails para MeOut
func ToMeOut(user *types.UserWithDetails) *MeOut {
	if user == nil {
		return nil
	}

	var addrOut *address.AddressOut
	if user.Address != nil {
		addrOut = &address.AddressOut{
			ID:           user.Address.ID,
			ZipCode:      user.Address.ZipCode,
			State:        user.Address.State,
			City:         user.Address.City,
			Neighborhood: user.Address.Neighborhood,
			Street:       user.Address.Street,
			Num:          user.Address.Num,
			Latitude:     user.Address.Latitude,
			Longitude:    user.Address.Longitude,
			AddInfo:      user.Address.AddInfo,
		}
	}

	var enterpriseOut *EnterpriseOut
	if user.Enterprise != nil {
		var enterpriseAddrOut *address.AddressOut
		if user.Enterprise.Address != nil {
			enterpriseAddrOut = &address.AddressOut{
				ID:           user.Enterprise.Address.ID,
				ZipCode:      user.Enterprise.Address.ZipCode,
				State:        user.Enterprise.Address.State,
				City:         user.Enterprise.Address.City,
				Neighborhood: user.Enterprise.Address.Neighborhood,
				Street:       user.Enterprise.Address.Street,
				Num:          user.Enterprise.Address.Num,
				Latitude:     user.Enterprise.Address.Latitude,
				Longitude:    user.Enterprise.Address.Longitude,
				AddInfo:      user.Enterprise.Address.AddInfo,
			}
		}

		enterpriseOut = &EnterpriseOut{
			ID:          user.Enterprise.ID,
			Name:        user.Enterprise.Name,
			CNPJ:        user.Enterprise.CNPJ,
			Email:       user.Enterprise.Email,
			FantasyName: user.Enterprise.FantasyName,
			Phone:       user.Enterprise.Phone,
			AddressID:   user.Enterprise.AddressID,
			Address:     enterpriseAddrOut,
		}
	}

	var roleTitle *string
	if user.Role != nil {
		roleTitle = &user.Role.Title
	}

	return &MeOut{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		Document:     user.Document,
		Phone:        user.Phone,
		Address:      addrOut,
		RoleTitle:    roleTitle,
		EnterpriseID: user.EnterpriseID,
		Enterprise:   enterpriseOut,
		Permissions:  ToPermissionOutSlice(user.Permissions),
	}
}

// ToPermissionOut converte um UserPermission para PermissionOut
func ToPermissionOut(p *types.UserPermission) *PermissionOut {
	if p == nil {
		return nil
	}

	return &PermissionOut{
		ID:          p.ID,
		FeatureID:   p.FeatureID,
		FeatureName: p.FeatureName,
		Create:      p.Create,
		Read:        p.Read,
		Update:      p.Update,
		Delete:      p.Delete,
	}
}

// ToPermissionOutSlice converte um slice de UserPermission para PermissionOut
func ToPermissionOutSlice(permissions []*types.UserPermission) []*PermissionOut {
	if permissions == nil {
		return nil
	}

	result := make([]*PermissionOut, len(permissions))
	for i, p := range permissions {
		result[i] = ToPermissionOut(p)
	}
	return result
}

// ToMyPermissionsOut converte um UserPermissions para MyPermissionsOut
func ToMyPermissionsOut(permissions *types.UserPermissions) *MyPermissionsOut {
	if permissions == nil {
		return nil
	}

	return &MyPermissionsOut{
		ID:          permissions.ID,
		Name:        permissions.Name,
		RoleTitle:   permissions.RoleTitle,
		Permissions: ToPermissionOutSlice(permissions.Permissions),
	}
}
