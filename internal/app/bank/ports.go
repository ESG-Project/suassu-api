package bank

import (
	"context"
	"encoding/json"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainbank "github.com/ESG-Project/suassu-api/internal/domain/bank"
)

// Repo cobre o catálogo global de bancos e os vínculos por empresa.
type Repo interface {
	GetByCode(ctx context.Context, code string) (*domainbank.Bank, error)
	Create(ctx context.Context, b *domainbank.Bank) error
	GetEnterpriseBank(ctx context.Context, bankID, enterpriseID string) (*domainbank.EnterpriseBank, error)
	GetEnterpriseBankByID(ctx context.Context, id string) (*types.EnterpriseBankDetail, error)
	CreateEnterpriseBank(ctx context.Context, eb *domainbank.EnterpriseBank) error
	DeleteEnterpriseBank(ctx context.Context, id string) error
	ListByEnterprise(ctx context.Context, enterpriseID string) ([]types.EnterpriseBankRow, error)
}

// Catalog é a fonte externa do catálogo público de bancos (GET /all-banks).
// A resposta é repassada verbatim ao cliente, então não há tipagem aqui.
type Catalog interface {
	List(ctx context.Context) (json.RawMessage, error)
}
