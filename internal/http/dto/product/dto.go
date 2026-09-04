package productdto

import (
	appproduct "github.com/ESG-Project/suassu-api/internal/app/product"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainproduct "github.com/ESG-Project/suassu-api/internal/domain/product"
	"github.com/ESG-Project/suassu-api/internal/http/dto/scalar"
)

// CreateProductRequest é o corpo de POST /product.
type CreateProductRequest struct {
	Name           string       `json:"name"`
	SuggestedValue *scalar.Text `json:"suggestedValue"`
	TypeProductID  *string      `json:"typeProductId"`
	Deliverable    bool         `json:"deliverable"`
	IsDefault      *bool        `json:"isDefault"`
}

func (r CreateProductRequest) ToInput() appproduct.CreateInput {
	return appproduct.CreateInput{
		Name:           r.Name,
		SuggestedValue: r.SuggestedValue.StringPtr(),
		TypeProductID:  r.TypeProductID,
		Deliverable:    r.Deliverable,
		IsDefault:      r.IsDefault,
	}
}

// UpdateProductRequest é o corpo de PUT /product — o id vem no corpo, não na
// URL, como no user-crud. Campos ausentes ficam nil e o serviço preserva o
// valor atual.
type UpdateProductRequest struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	SuggestedValue *scalar.Text `json:"suggestedValue"`
	TypeProductID  *string      `json:"typeProductId"`
	Deliverable    *bool        `json:"deliverable"`
	IsDefault      *bool        `json:"isDefault"`
}

func (r UpdateProductRequest) ToInput() appproduct.UpdateInput {
	return appproduct.UpdateInput{
		ID:             r.ID,
		Name:           r.Name,
		SuggestedValue: r.SuggestedValue.StringPtr(),
		TypeProductID:  r.TypeProductID,
		Deliverable:    r.Deliverable,
		IsDefault:      r.IsDefault,
	}
}

// ProductResponse é a linha de Product crua, como o Prisma devolvia em
// POST/PUT /product.
type ProductResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	SuggestedValue *string `json:"suggestedValue"`
	EnterpriseID   string  `json:"enterpriseId"`
	ParameterID    *string `json:"parameterId"`
	Deliverable    bool    `json:"deliverable"`
	TypeProductID  *string `json:"typeProductId"`
	IsDefault      bool    `json:"isDefault"`
}

func ToProductResponse(p *domainproduct.Product) ProductResponse {
	return ProductResponse{
		ID:             p.ID,
		Name:           p.Name,
		SuggestedValue: p.SuggestedValue,
		EnterpriseID:   p.EnterpriseID,
		ParameterID:    p.ParameterID,
		Deliverable:    p.Deliverable,
		TypeProductID:  p.TypeProductID,
		IsDefault:      p.IsDefault,
	}
}

// ParameterRef é o parâmetro aninhado em cada item de GET /product/enterprise.
// A chave "Parameter" vem em maiúscula por herança do select do Prisma.
type ParameterRef struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Value *string `json:"value"`
}

// TypeRef é o tipo de produto aninhado em cada item de GET /product/enterprise.
type TypeRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ProductListItem é um item de GET /product/enterprise. O conjunto de campos
// (sem parameterId/typeProductId, com Parameter e type aninhados) é o mesmo
// que o select do user-crud devolvia.
type ProductListItem struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	SuggestedValue *string       `json:"suggestedValue"`
	EnterpriseID   string        `json:"enterpriseId"`
	Parameter      *ParameterRef `json:"Parameter"`
	Deliverable    bool          `json:"deliverable"`
	Type           *TypeRef      `json:"type"`
}

func ToProductListItems(rows []types.ProductDetailRow) []ProductListItem {
	out := make([]ProductListItem, 0, len(rows))
	for _, row := range rows {
		item := ProductListItem{
			ID:             row.ID,
			Name:           row.Name,
			SuggestedValue: row.SuggestedValue,
			EnterpriseID:   row.EnterpriseID,
			Deliverable:    row.Deliverable,
		}
		if row.Parameter != nil {
			item.Parameter = &ParameterRef{ID: row.Parameter.ID, Title: row.Parameter.Title, Value: row.Parameter.Value}
		}
		if row.Type != nil {
			item.Type = &TypeRef{ID: row.Type.ID, Type: row.Type.Type}
		}
		out = append(out, item)
	}
	return out
}
