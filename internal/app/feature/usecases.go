package feature

import (
	"context"
	"log"
	"strings"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainfeature "github.com/ESG-Project/suassu-api/internal/domain/feature"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/seeds"
	"github.com/google/uuid"
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

// List retorna todas as features cadastradas.
func (s *Service) List(ctx context.Context) ([]domainfeature.Feature, error) {
	return s.repo.List(ctx)
}

// Create cadastra uma nova feature (uso administrativo).
func (s *Service) Create(ctx context.Context, name string) (*domainfeature.Feature, error) {
	if strings.TrimSpace(name) == "" {
		return nil, apperr.New(apperr.CodeInvalid, "name is required")
	}

	f := domainfeature.NewFeature(uuid.NewString(), name)
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

// Delete remove uma feature (uso administrativo).
func (s *Service) Delete(ctx context.Context, id string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperr.New(apperr.CodeNotFound, "feature not found")
	}
	return s.repo.Delete(ctx, id)
}
