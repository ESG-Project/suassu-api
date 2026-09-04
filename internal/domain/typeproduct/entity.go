package typeproduct

import (
	"errors"
	"strings"
)

// TypeProduct é a categoria de um produto dentro da empresa (tabela
// "TypeProduct"). O campo Type mantém o nome herdado do Prisma.
type TypeProduct struct {
	ID           string
	Type         string
	EnterpriseID string
}

func NewTypeProduct(id, typ, enterpriseID string) *TypeProduct {
	return &TypeProduct{ID: id, Type: typ, EnterpriseID: enterpriseID}
}

func (t *TypeProduct) Validate() error {
	if strings.TrimSpace(t.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(t.Type) == "" {
		return errors.New("type is required")
	}
	if strings.TrimSpace(t.EnterpriseID) == "" {
		return errors.New("enterpriseId is required")
	}
	return nil
}

func (t *TypeProduct) SetType(typ string) { t.Type = typ }
