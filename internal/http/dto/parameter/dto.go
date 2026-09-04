package parameterdto

import (
	appparameter "github.com/ESG-Project/suassu-api/internal/app/parameter"
	domainparameter "github.com/ESG-Project/suassu-api/internal/domain/parameter"
	"github.com/ESG-Project/suassu-api/internal/http/dto/scalar"
)

// UpdateParameterRequest é o corpo de PUT /parameter — o id vem no corpo, não
// na URL, como no user-crud.
//
// Value chega como string ou número (o front manda o valor do formulário sem
// converter); scalar.Text normaliza os dois. Campos ausentes ficam nil e o
// serviço preserva o valor atual.
type UpdateParameterRequest struct {
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	Value     *scalar.Text `json:"value"`
	IsDefault *bool        `json:"isDefault"`
}

func (r UpdateParameterRequest) ToInput() appparameter.UpdateInput {
	return appparameter.UpdateInput{
		ID:        r.ID,
		Title:     r.Title,
		Value:     r.Value.StringPtr(),
		IsDefault: r.IsDefault,
	}
}

// ParameterResponse é a linha de Parameter crua, como o Prisma devolvia.
type ParameterResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Value        *string `json:"value"`
	EnterpriseID string  `json:"enterpriseId"`
	IsDefault    bool    `json:"isDefault"`
}

func ToParameterResponse(p *domainparameter.Parameter) ParameterResponse {
	return ParameterResponse{
		ID:           p.ID,
		Title:        p.Title,
		Value:        p.Value,
		EnterpriseID: p.EnterpriseID,
		IsDefault:    p.IsDefault,
	}
}

func ToParameterResponses(list []domainparameter.Parameter) []ParameterResponse {
	out := make([]ParameterResponse, 0, len(list))
	for i := range list {
		out = append(out, ToParameterResponse(&list[i]))
	}
	return out
}
