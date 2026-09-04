package postgres

import (
	"context"
	"database/sql"

	domaintypeproduct "github.com/ESG-Project/suassu-api/internal/domain/typeproduct"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type TypeProductRepo struct{ q *sqlc.Queries }

func NewTypeProductRepoFrom(d dbtx) *TypeProductRepo { return &TypeProductRepo{q: sqlc.New(d)} }
func NewTypeProductRepo(db *sql.DB) *TypeProductRepo { return NewTypeProductRepoFrom(db) }

func (r *TypeProductRepo) Create(ctx context.Context, t *domaintypeproduct.TypeProduct) error {
	return r.q.CreateTypeProduct(ctx, sqlc.CreateTypeProductParams{
		ID:           t.ID,
		Type:         t.Type,
		EnterpriseId: t.EnterpriseID,
	})
}

// GetByID busca sem filtrar por empresa: quem chama compara o enterpriseId
// para distinguir 404 de 403, como faz o user-crud.
func (r *TypeProductRepo) GetByID(ctx context.Context, id string) (*domaintypeproduct.TypeProduct, error) {
	row, err := r.q.GetTypeProductByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &domaintypeproduct.TypeProduct{ID: row.ID, Type: row.Type, EnterpriseID: row.EnterpriseId}, nil
}

func (r *TypeProductRepo) List(ctx context.Context, enterpriseID string) ([]domaintypeproduct.TypeProduct, error) {
	rows, err := r.q.ListTypeProductsByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}
	out := make([]domaintypeproduct.TypeProduct, 0, len(rows))
	for _, row := range rows {
		out = append(out, domaintypeproduct.TypeProduct{ID: row.ID, Type: row.Type, EnterpriseID: row.EnterpriseId})
	}
	return out, nil
}

func (r *TypeProductRepo) Update(ctx context.Context, t *domaintypeproduct.TypeProduct) error {
	return r.q.UpdateTypeProduct(ctx, sqlc.UpdateTypeProductParams{
		ID:           t.ID,
		Type:         t.Type,
		EnterpriseId: t.EnterpriseID,
	})
}

func (r *TypeProductRepo) Delete(ctx context.Context, id, enterpriseID string) error {
	return r.q.DeleteTypeProduct(ctx, sqlc.DeleteTypeProductParams{ID: id, EnterpriseId: enterpriseID})
}
