package roledto

import (
	domainrole "github.com/ESG-Project/suassu-api/internal/domain/role"
)

// CreateRoleRequest representa o corpo para criar um novo papel.
type CreateRoleRequest struct {
	Title string `json:"title"`
}

// RoleResponse representa a resposta de um papel (mesmo shape do user-crud:
// id, title, enterpriseId).
type RoleResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	EnterpriseID string `json:"enterpriseId"`
}

func ToRoleResponse(r *domainrole.Role) *RoleResponse {
	return &RoleResponse{ID: r.ID, Title: r.Title, EnterpriseID: r.EnterpriseID}
}

func ToRoleResponses(list []domainrole.Role) []*RoleResponse {
	out := make([]*RoleResponse, 0, len(list))
	for i := range list {
		out = append(out, ToRoleResponse(&list[i]))
	}
	return out
}
