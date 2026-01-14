package specimen

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainspecimen "github.com/ESG-Project/suassu-api/internal/domain/specimen"
)

// Repo define a interface do repositório de Specimen
type Repo interface {
	Create(ctx context.Context, s *domainspecimen.Specimen) error
	GetByID(ctx context.Context, id string) (*types.SpecimenWithSpecies, error)
	ListByPhytoAnalysis(ctx context.Context, phytoAnalysisID string) ([]*types.SpecimenWithSpecies, error)
	Update(ctx context.Context, s *domainspecimen.Specimen) error
	Delete(ctx context.Context, id string) error
	CountByPhytoAnalysis(ctx context.Context, phytoAnalysisID string) (int64, error)
}

