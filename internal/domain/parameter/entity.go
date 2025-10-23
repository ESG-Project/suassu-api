package parameter

import (
	"errors"
	"strings"
)

type Parameter struct {
	ID           string
	Title        string
	Value        *string
	EnterpriseID string
	IsDefault    bool
}

func NewParameter(id, title, enterpriseID string) *Parameter {
	return &Parameter{
		ID:           id,
		Title:        title,
		EnterpriseID: enterpriseID,
		IsDefault:    false,
	}
}

func (p *Parameter) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(p.EnterpriseID) == "" {
		return errors.New("enterpriseId is required")
	}
	return nil
}

func (p *Parameter) SetValue(value *string) {
	p.Value = value
}

func (p *Parameter) SetIsDefault(isDefault bool) {
	p.IsDefault = isDefault
}
