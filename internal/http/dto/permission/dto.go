package permissiondto

import domainpermission "github.com/ESG-Project/suassu-api/internal/domain/permission"

// CreatePermissionRequest representa o corpo para criar uma permissão.
// O campo de exclusão chama-se "erase" no JSON por compatibilidade com o
// contrato herdado do user-crud (delete é reservado em alguns contextos lá).
type CreatePermissionRequest struct {
	FeatureID string `json:"featureId"`
	RoleID    string `json:"roleId"`
	Create    bool   `json:"create"`
	Read      bool   `json:"read"`
	Update    bool   `json:"update"`
	Erase     bool   `json:"erase"`
}

// UpdatePermissionRequest representa o corpo para atualizar uma permissão.
type UpdatePermissionRequest struct {
	ID        string `json:"id"`
	FeatureID string `json:"featureId"`
	RoleID    string `json:"roleId"`
	Create    bool   `json:"create"`
	Read      bool   `json:"read"`
	Update    bool   `json:"update"`
	Erase     bool   `json:"erase"`
}

// PermissionResponse representa a resposta de uma permissão (mesmo shape do
// user-crud: id, featureId, roleId, create, read, update, delete).
type PermissionResponse struct {
	ID        string `json:"id"`
	FeatureID string `json:"featureId"`
	RoleID    string `json:"roleId"`
	Create    bool   `json:"create"`
	Read      bool   `json:"read"`
	Update    bool   `json:"update"`
	Delete    bool   `json:"delete"`
}

func ToPermissionResponse(p *domainpermission.Permission) *PermissionResponse {
	return &PermissionResponse{
		ID:        p.ID,
		FeatureID: p.FeatureID,
		RoleID:    p.RoleID,
		Create:    p.Create,
		Read:      p.Read,
		Update:    p.Update,
		Delete:    p.Delete,
	}
}

func ToPermissionResponses(list []domainpermission.Permission) []*PermissionResponse {
	out := make([]*PermissionResponse, 0, len(list))
	for i := range list {
		out = append(out, ToPermissionResponse(&list[i]))
	}
	return out
}
