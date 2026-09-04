package product

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainproduct "github.com/ESG-Project/suassu-api/internal/domain/product"
)

type Repo interface {
	Create(ctx context.Context, p *domainproduct.Product) error
	GetByIDAnyEnterprise(ctx context.Context, id string) (*domainproduct.Product, error)
	ListDetailedByEnterprise(ctx context.Context, enterpriseID string) ([]types.ProductDetailRow, error)
	Update(ctx context.Context, p *domainproduct.Product) error
	Delete(ctx context.Context, id, enterpriseID string) error
}
