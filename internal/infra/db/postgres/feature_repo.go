package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/domain/feature"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type FeatureRepo struct{ q *sqlc.Queries }

// NewFeatureRepoFrom accepts any sqlc.DBTX (e.g. *sql.DB or *sql.Tx)
func NewFeatureRepoFrom(d dbtx) *FeatureRepo { return &FeatureRepo{q: sqlc.New(d)} }

// NewFeatureRepo for backward compatibility
func NewFeatureRepo(db *sql.DB) *FeatureRepo { return NewFeatureRepoFrom(db) }

// Upsert insere uma nova feature ou não faz nada se o nome já existir.
func (r *FeatureRepo) Upsert(ctx context.Context, name string) error {
	return r.q.UpsertFeature(ctx, name)
}

// List retorna todas as features cadastradas.
func (r *FeatureRepo) List(ctx context.Context) ([]feature.Feature, error) {
	rows, err := r.q.ListAllFeatures(ctx)
	if err != nil {
		return nil, err
	}

	features := make([]feature.Feature, 0, len(rows))
	for _, row := range rows {
		features = append(features, feature.Feature{
			ID:   row.ID,
			Name: row.Name,
		})
	}
	return features, nil
}

// GetByName retorna uma feature pelo nome.
func (r *FeatureRepo) GetByName(ctx context.Context, name string) (*feature.Feature, error) {
	row, err := r.q.GetFeatureByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &feature.Feature{
		ID:   row.ID,
		Name: row.Name,
	}, nil
}
