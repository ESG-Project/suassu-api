package typeproduct

import (
	"context"

	domaintypeproduct "github.com/ESG-Project/suassu-api/internal/domain/typeproduct"
)

type Repo interface {
	Create(ctx context.Context, t *domaintypeproduct.TypeProduct) error
	GetByID(ctx context.Context, id string) (*domaintypeproduct.TypeProduct, error)
	List(ctx context.Context, enterpriseID string) ([]domaintypeproduct.TypeProduct, error)
	Update(ctx context.Context, t *domaintypeproduct.TypeProduct) error
	Delete(ctx context.Context, id, enterpriseID string) error
}
