package typeproductdto

import (
	domaintypeproduct "github.com/ESG-Project/suassu-api/internal/domain/typeproduct"
)

// CreateTypeProductRequest é o corpo de POST /typeProduct.
type CreateTypeProductRequest struct {
	Type string `json:"type"`
}

// UpdateTypeProductRequest é o corpo de PUT /typeProduct — o id vem no corpo,
// não na URL, como no user-crud.
type UpdateTypeProductRequest struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// TypeProductResponse é a linha de TypeProduct crua, como o Prisma devolvia.
type TypeProductResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	EnterpriseID string `json:"enterpriseId"`
}

func ToTypeProductResponse(t *domaintypeproduct.TypeProduct) TypeProductResponse {
	return TypeProductResponse{ID: t.ID, Type: t.Type, EnterpriseID: t.EnterpriseID}
}

func ToTypeProductResponses(list []domaintypeproduct.TypeProduct) []TypeProductResponse {
	out := make([]TypeProductResponse, 0, len(list))
	for i := range list {
		out = append(out, ToTypeProductResponse(&list[i]))
	}
	return out
}
