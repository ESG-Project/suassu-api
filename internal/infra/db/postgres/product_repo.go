package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainproduct "github.com/ESG-Project/suassu-api/internal/domain/product"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type ProductRepo struct{ q *sqlc.Queries }

func NewProductRepoFrom(d dbtx) *ProductRepo {
	return &ProductRepo{q: sqlc.New(d)}
}

func NewProductRepo(db *sql.DB) *ProductRepo {
	return &ProductRepo{q: sqlc.New(db)}
}

func (r *ProductRepo) Create(ctx context.Context, p *domainproduct.Product) error {
	return r.q.CreateProduct(ctx, sqlc.CreateProductParams{
		ID:             p.ID,
		Name:           p.Name,
		SuggestedValue: utils.ToNullString(p.SuggestedValue),
		EnterpriseId:   p.EnterpriseID,
		ParameterId:    utils.ToNullString(p.ParameterID),
		Deliverable:    p.Deliverable,
		TypeProductId:  utils.ToNullString(p.TypeProductID),
		IsDefault:      p.IsDefault,
	})
}

func (r *ProductRepo) GetByID(ctx context.Context, id, enterpriseID string) (*domainproduct.Product, error) {
	row, err := r.q.GetProductByID(ctx, sqlc.GetProductByIDParams{
		ID:           id,
		EnterpriseId: enterpriseID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "product not found")
		}
		return nil, err
	}

	return &domainproduct.Product{
		ID:             row.ID,
		Name:           row.Name,
		SuggestedValue: utils.FromNullString(row.SuggestedValue),
		EnterpriseID:   row.EnterpriseId,
		ParameterID:    utils.FromNullString(row.ParameterId),
		Deliverable:    row.Deliverable,
		TypeProductID:  utils.FromNullString(row.TypeProductId),
		IsDefault:      row.IsDefault,
	}, nil
}

func (r *ProductRepo) List(ctx context.Context, enterpriseID string) ([]domainproduct.Product, error) {
	rows, err := r.q.ListProductsByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	products := make([]domainproduct.Product, 0, len(rows))
	for _, row := range rows {
		products = append(products, domainproduct.Product{
			ID:             row.ID,
			Name:           row.Name,
			SuggestedValue: utils.FromNullString(row.SuggestedValue),
			EnterpriseID:   row.EnterpriseId,
			ParameterID:    utils.FromNullString(row.ParameterId),
			Deliverable:    row.Deliverable,
			TypeProductID:  utils.FromNullString(row.TypeProductId),
			IsDefault:      row.IsDefault,
		})
	}
	return products, nil
}

func (r *ProductRepo) Update(ctx context.Context, p *domainproduct.Product) error {
	return r.q.UpdateProduct(ctx, sqlc.UpdateProductParams{
		ID:             p.ID,
		Name:           p.Name,
		SuggestedValue: utils.ToNullString(p.SuggestedValue),
		ParameterId:    utils.ToNullString(p.ParameterID),
		Deliverable:    p.Deliverable,
		TypeProductId:  utils.ToNullString(p.TypeProductID),
		IsDefault:      p.IsDefault,
		EnterpriseId:   p.EnterpriseID,
	})
}

func (r *ProductRepo) Delete(ctx context.Context, id, enterpriseID string) error {
	return r.q.DeleteProduct(ctx, sqlc.DeleteProductParams{
		ID:           id,
		EnterpriseId: enterpriseID,
	})
}
