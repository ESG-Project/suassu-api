package species

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	domainspecies "github.com/ESG-Project/suassu-api/internal/domain/species"
)

// Repo define a interface do repositório de Species
type Repo interface {
	CreateLegislation(ctx context.Context, sl *domainspecies.SpeciesLegislation) error
	CreateSpecies(ctx context.Context, s *domainspecies.Species) error
	GetByID(ctx context.Context, id string) (*types.SpeciesWithLegislation, error)
	GetByScientificName(ctx context.Context, scientificName string) (*types.SpeciesWithLegislation, error)
	GetMapByScientificNames(ctx context.Context, names []string) (map[string]string, error)
	List(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error)
	UpdateSpecies(ctx context.Context, s *domainspecies.Species) error
	UpdateLegislation(ctx context.Context, sl *domainspecies.SpeciesLegislation) error
}

