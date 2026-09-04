package bankdto

import (
	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainbank "github.com/ESG-Project/suassu-api/internal/domain/bank"
)

// CreateBankRequest é o corpo de POST /bank — o banco vem aninhado em "bank",
// como no user-crud.
type CreateBankRequest struct {
	Bank struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"bank"`
}

// EnterpriseBankResponse é a resposta de POST /bank: a linha de
// EnterpriseBank crua, como o Prisma devolvia.
type EnterpriseBankResponse struct {
	ID           string `json:"id"`
	EnterpriseID string `json:"enterpriseId"`
	BankID       string `json:"bankId"`
}

func ToEnterpriseBankResponse(eb *domainbank.EnterpriseBank) EnterpriseBankResponse {
	return EnterpriseBankResponse{ID: eb.ID, EnterpriseID: eb.EnterpriseID, BankID: eb.BankID}
}

// BankRef é o banco aninhado em cada item de GET /bank/enterprise.
type BankRef struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// EnterpriseBankListItem é um item de GET /bank/enterprise.
type EnterpriseBankListItem struct {
	ID   string  `json:"id"`
	Bank BankRef `json:"bank"`
}

func ToEnterpriseBankListItems(rows []types.EnterpriseBankRow) []EnterpriseBankListItem {
	out := make([]EnterpriseBankListItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, EnterpriseBankListItem{
			ID:   row.ID,
			Bank: BankRef{ID: row.BankID, Code: row.BankCode, Name: row.BankName},
		})
	}
	return out
}
