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
	GetNextVersion(ctx context.Context, scientificName string) (int32, error)
	CountLineage(ctx context.Context, speciesID string) (int32, error)
	List(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error)
	ListVisible(ctx context.Context, enterpriseID string, limit, offset int32) ([]*types.SpeciesWithLegislation, error)
	ListPending(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error)
	UpdateSpecies(ctx context.Context, s *domainspecies.Species) error
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateLegislation(ctx context.Context, sl *domainspecies.SpeciesLegislation) error
}

