package authhttp

import (
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/http/dto"
)

// Converter de tipos de aplicação para DTOs HTTP
func convertUserWithDetailsToMeOut(user *types.UserWithDetails) *dto.MeOut {
	var address *dto.AddressOut
	if user.Address != nil {
		address = &dto.AddressOut{
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

	var enterprise *dto.EnterpriseOut
	if user.Enterprise != nil {
		enterprise = &dto.EnterpriseOut{
			ID:          user.Enterprise.ID,
			Name:        user.Enterprise.Name,
			CNPJ:        user.Enterprise.CNPJ,
			Email:       user.Enterprise.Email,
			FantasyName: user.Enterprise.FantasyName,
			Phone:       user.Enterprise.Phone,
			AddressID:   user.Enterprise.AddressID,
		}
	}

	var roleTitle *string
	if user.Role != nil {
		roleTitle = &user.Role.Title
	}

	return &dto.MeOut{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		Document:     user.Document,
		Phone:        user.Phone,
		Address:      address,
		RoleTitle:    roleTitle,
		EnterpriseID: user.EnterpriseID,
		Enterprise:   enterprise,
		Permissions:  convertPermissionsToDTO(user.Permissions),
	}
}

func convertUserPermissionsToMyPermissionsOut(permissions *types.UserPermissions) *dto.MyPermissionsOut {
	return &dto.MyPermissionsOut{
		ID:          permissions.ID,
		Name:        permissions.Name,
		RoleTitle:   permissions.RoleTitle,
		Permissions: convertPermissionsToDTO(permissions.Permissions),
	}
}

func convertPermissionsToDTO(permissions []*types.UserPermission) []*dto.PermissionOut {
	if permissions == nil {
		return nil
	}

	result := make([]*dto.PermissionOut, len(permissions))
	for i, p := range permissions {
		result[i] = &dto.PermissionOut{
			ID:          p.ID,
			FeatureID:   p.FeatureID,
			FeatureName: p.FeatureName,
			Create:      p.Create,
			Read:        p.Read,
			Update:      p.Update,
			Delete:      p.Delete,
		}
	}
	return result
}
