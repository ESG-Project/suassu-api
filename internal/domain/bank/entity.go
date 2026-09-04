package bank

import (
	"errors"
	"strings"
)

// Bank é um banco do catálogo global (tabela "Bank"): code e name são únicos
// e compartilhados por todas as empresas.
type Bank struct {
	ID   string
	Code string
	Name string
}

func NewBank(id, code, name string) *Bank {
	return &Bank{ID: id, Code: code, Name: name}
}

func (b *Bank) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(b.Code) == "" {
		return errors.New("code is required")
	}
	if strings.TrimSpace(b.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}

// EnterpriseBank é o vínculo entre uma empresa e um banco do catálogo.
type EnterpriseBank struct {
	ID           string
	EnterpriseID string
	BankID       string
}

func NewEnterpriseBank(id, enterpriseID, bankID string) *EnterpriseBank {
	return &EnterpriseBank{ID: id, EnterpriseID: enterpriseID, BankID: bankID}
}

func (e *EnterpriseBank) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(e.EnterpriseID) == "" {
		return errors.New("enterpriseId is required")
	}
	if strings.TrimSpace(e.BankID) == "" {
		return errors.New("bankId is required")
	}
	return nil
}
