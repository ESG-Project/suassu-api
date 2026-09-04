package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainbank "github.com/ESG-Project/suassu-api/internal/domain/bank"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type BankRepo struct{ q *sqlc.Queries }

func NewBankRepoFrom(d dbtx) *BankRepo { return &BankRepo{q: sqlc.New(d)} }
func NewBankRepo(db *sql.DB) *BankRepo { return NewBankRepoFrom(db) }

// GetByCode devolve nil (sem erro) quando o banco ainda não está no catálogo.
func (r *BankRepo) GetByCode(ctx context.Context, code string) (*domainbank.Bank, error) {
	row, err := r.q.GetBankByCode(ctx, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &domainbank.Bank{ID: row.ID, Code: row.Code, Name: row.Name}, nil
}

func (r *BankRepo) Create(ctx context.Context, b *domainbank.Bank) error {
	return r.q.CreateBank(ctx, sqlc.CreateBankParams{ID: b.ID, Code: b.Code, Name: b.Name})
}

// GetEnterpriseBank devolve nil quando o vínculo ainda não existe.
func (r *BankRepo) GetEnterpriseBank(ctx context.Context, bankID, enterpriseID string) (*domainbank.EnterpriseBank, error) {
	row, err := r.q.GetEnterpriseBankByBankAndEnterprise(ctx, sqlc.GetEnterpriseBankByBankAndEnterpriseParams{
		BankId:       bankID,
		EnterpriseId: enterpriseID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &domainbank.EnterpriseBank{ID: row.ID, EnterpriseID: row.EnterpriseId, BankID: row.BankId}, nil
}

// GetEnterpriseBankByID busca sem filtrar por empresa, para que o serviço
// possa distinguir "não existe" (404) de "é de outra empresa" (403), como faz
// o DeleteBankService do user-crud.
func (r *BankRepo) GetEnterpriseBankByID(ctx context.Context, id string) (*types.EnterpriseBankDetail, error) {
	row, err := r.q.GetEnterpriseBankByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &types.EnterpriseBankDetail{
		ID:             row.ID,
		EnterpriseID:   row.EnterpriseId,
		BankID:         row.BankId,
		EnterpriseName: row.EnterpriseName,
	}, nil
}

func (r *BankRepo) CreateEnterpriseBank(ctx context.Context, eb *domainbank.EnterpriseBank) error {
	return r.q.CreateEnterpriseBank(ctx, sqlc.CreateEnterpriseBankParams{
		ID:           eb.ID,
		EnterpriseId: eb.EnterpriseID,
		BankId:       eb.BankID,
	})
}

func (r *BankRepo) DeleteEnterpriseBank(ctx context.Context, id string) error {
	return r.q.DeleteEnterpriseBank(ctx, id)
}

func (r *BankRepo) ListByEnterprise(ctx context.Context, enterpriseID string) ([]types.EnterpriseBankRow, error) {
	rows, err := r.q.ListEnterpriseBanksByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}
	out := make([]types.EnterpriseBankRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, types.EnterpriseBankRow{
			ID:       row.ID,
			BankID:   row.BankID,
			BankCode: row.BankCode,
			BankName: row.BankName,
		})
	}
	return out, nil
}
