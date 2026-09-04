package parameter

import (
	"context"

	domainparameter "github.com/ESG-Project/suassu-api/internal/domain/parameter"
)

type Repo interface {
	GetByIDAnyEnterprise(ctx context.Context, id string) (*domainparameter.Parameter, error)
	List(ctx context.Context, enterpriseID string) ([]domainparameter.Parameter, error)
	Update(ctx context.Context, p *domainparameter.Parameter) error
}
