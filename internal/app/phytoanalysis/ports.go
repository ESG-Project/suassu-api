package phytoanalysis

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainphyto "github.com/ESG-Project/suassu-api/internal/domain/phytoanalysis"
)

// Repo define a interface do repositório de PhytoAnalysis
type Repo interface {
	Create(ctx context.Context, p *domainphyto.PhytoAnalysis) error
	GetByID(ctx context.Context, id string) (*types.PhytoAnalysisWithProject, error)
	ListByProject(ctx context.Context, projectID string) ([]*types.PhytoAnalysisWithProject, error)
	ListAll(ctx context.Context, limit, offset int32) ([]*types.PhytoAnalysisWithProject, error)
	Update(ctx context.Context, p *domainphyto.PhytoAnalysis) error
	Delete(ctx context.Context, id string) error
	GetWithSpecimens(ctx context.Context, id string) (*types.PhytoAnalysisComplete, error)
}
