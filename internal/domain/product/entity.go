package product

import (
	"errors"
	"strings"
)

type Product struct {
	ID             string
	Name           string
	SuggestedValue *string
	EnterpriseID   string
	ParameterID    *string
	Deliverable    bool
	TypeProductID  *string
	IsDefault      bool
}

func NewProduct(id, name, enterpriseID string, deliverable bool) *Product {
	return &Product{
		ID:           id,
		Name:         name,
		EnterpriseID: enterpriseID,
		Deliverable:  deliverable,
		IsDefault:    false,
	}
}

func (p *Product) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(p.EnterpriseID) == "" {
		return errors.New("enterpriseId is required")
	}
	return nil
}

func (p *Product) SetSuggestedValue(value *string) {
	p.SuggestedValue = value
}

func (p *Product) SetParameterID(parameterID *string) {
	p.ParameterID = parameterID
}

func (p *Product) SetTypeProductID(typeProductID *string) {
	p.TypeProductID = typeProductID
}

func (p *Product) SetIsDefault(isDefault bool) {
	p.IsDefault = isDefault
}
