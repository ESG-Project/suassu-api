package features

import (
	"context"
	"log"

	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/seeds"
)

type Service struct {
	repo   Repo
	hasher Hasher
	txm    postgres.TxManagerInterface
}

func NewService(r Repo, h Hasher) *Service {
	return NewServiceWithTx(r, h, nil)
}

func NewServiceWithTx(r Repo, h Hasher, txm postgres.TxManagerInterface) *Service {
	return &Service{repo: r, hasher: h, txm: txm}
}

type CreateInput struct {
	Name string
}

// SeedFeatures itera sobre a lista de features predefinida e as insere no banco.
func (s *Service) SeedFeatures(ctx context.Context) {
	log.Println("Checking and populating features table...")

	for _, featureName := range seeds.FeatureList {
		if err := s.repo.Upsert(ctx, featureName); err != nil {
			log.Printf("failed to upsert feature '%s': %v", featureName, err)
		}
	}

	log.Println("Features table is up to date.")
}
